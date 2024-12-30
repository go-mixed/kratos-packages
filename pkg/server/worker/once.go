package worker

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/schedule"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/task"
	"time"
)

type onceWorker struct {
	key string
	// worker 外部创建时，必须是worker的clone体。因为会修改worker.store的属性
	worker *Worker
	// override the key
	override bool
}

var _ IWorker = (*onceWorker)(nil)

func (w *onceWorker) runScript(ctx context.Context, code string, keys []string, args ...any) *redis.Cmd {
	script := w.worker.store.Script(code)
	return script.Run(ctx, keys, args...)
}

const onceJobScript = `
local key = KEYS[1]
local app_id = ARGV[1]
local expiration = tonumber(ARGV[2]) or 0 -- 如果 ARGV[2] 不是数字，则设置为 0

local str = redis.call('GET', key)
local js = cjson.decode(str)
local now = redis.call('TIME')

if not str or not js or type(js) ~= 'table'  then
	js = {
		app_id = app_id
		created_at = now[1],
		last_refresh_at = now[1],
	}
end

if js['app_id'] == app_id then
	js['last_refresh_at'] = now[1]
	redis.call('SET', key, cjson.encode(js), 'PX', expiration)
	return true
end

return false
`

// wrapperOnceJobWithError 执行1次job：保证在集群中，这个job执行期间内，同名key绝对只会执行1次
func (w *onceWorker) wrapperOnceJobWithError(ctx context.Context, key string, interval time.Duration, job task.JobWithError) task.JobWithError {
	logger := w.worker.logger.WithContext(ctx)

	// 通过lua来创建key，并设置过期时间，确保本job在当前时间段内只会被执行1次
	ok, err := w.runScript(ctx, onceJobScript, []string{key}, w.worker.app.ID(), interval.Milliseconds()).Bool()
	if ok || err != nil {
		return func(ctx context.Context) error {
			// 创建一个ticker，用于给key续期
			refreshKeyTaskId := w.worker.WithContext(ctx).SubmitLoop(interval/2, func(ctx context.Context) {
				if err1 := w.runScript(ctx, onceJobScript, []string{key}, w.worker.app.ID(), interval.Milliseconds()).Err(); err1 != nil {
					logger.Errorf("[locker]renewal key %s failed: %v", key, err1)
				}
			})
			defer func() {
				// 取消刷新key的ticker，并删除key
				w.worker.CancelTask(refreshKeyTaskId)
				_, _ = w.worker.store.Del(ctx, key)
			}()

			return job(ctx)
		}
	}

	return func(ctx context.Context) error {
		return nil
	}
}

// wrapperOnceJob 执行1次job：保证在集群中，这个job执行期间内，同名key绝对只会执行1次
func (w *onceWorker) wrapperOnceJob(ctx context.Context, key string, refreshInterval time.Duration, job task.Job) task.Job {
	logger := w.worker.logger.WithContext(ctx)

	// 通过lua来创建key，并设置过期时间，确保本job在当前时间段内只会被执行1次
	ok, err := w.runScript(ctx, onceJobScript, []string{key}, w.worker.app.ID(), refreshInterval.Milliseconds()).Bool()
	if ok || err != nil {
		return func(ctx context.Context) {
			// 创建一个ticker，用于给key续期
			refreshKeyTaskId := w.worker.WithContext(ctx).SubmitLoop(refreshInterval/2, func(ctx context.Context) {
				if err1 := w.runScript(ctx, onceJobScript, []string{key}, w.worker.app.ID(), refreshInterval.Milliseconds()).Err(); err1 != nil {
					logger.Errorf("[locker]renewal key %s failed: %v", key, err1)
				}
			})
			defer func() {
				// 取消刷新key的ticker，并删除key
				w.worker.CancelTask(refreshKeyTaskId)
				_, _ = w.worker.store.Del(ctx, key)
			}()

			job(ctx)
		}
	}

	return func(ctx context.Context) {

	}
}

// cronOnceScript 保证整个集群中，每次cron触发时，该key的job只会执行一次。LUA可以保证原子性。
const cronOnceScript = `local key = KEYS[1]
local now = tonumber(ARGV[1])
local next_at = tonumber(ARGV[2])
local expiration = tonumber(ARGV[3])
local app_id = ARGV[4]
local last = redis.call('get', key)
local res = 0
local created_at = redis.call('TIME')[1]

if last then -- key 存在
	local js = cjson.decode(last)
	if js == nil then -- json不合法，可以执行
		res = 2
	elseif js['next_at'] == nil then -- next_at不存在，可以执行 
		res = 3
	elseif js['next_at'] > now then -- 未到执行时间，跳出
		return -1
	else -- 已经超过了执行时间，可以执行
		res = 4
	end
	
	-- 读取原来的created_at
	if js ~= nil and js['created_at'] ~= nil then
		created_at = js['created_at']
	end
else -- key不存在，可以执行
	res = 1
end


-- res > 0 表示可以执行，设置key和下次执行时间
redis.call('set', key, cjson.encode({last_at = now, next_at = next_at, app_id = app_id, created_at = created_at}), 'px', expiration)
return res
`

// wrapperOnceCronJob 保证整个集群中，每次cron触发时，该key的job只会执行一次
// （注意：基于的是每次cron触发时的时间点，比如任务是EveryMinute，那么表示每分钟在集群中只会执行一次）
func (w *onceWorker) wrapperOnceCronJob(key string, _cronSchedule *cronSchedule, job task.Job) task.Job {
	if key == "" || _cronSchedule == nil {
		return job
	}

	// 为了确保cron的多个节点的时间一致，这里计算出redis服务器时间与本地时间的差值，
	// 后面的now, nextTime都根据delta修正为redis服务器时间
	delta := w.worker.store.ServerTimeDelta(context.Background())

	// 录入cron任务时，如果key不存在，就设置下次运行的时间
	// 比如：程序滚动发布时，上一个执行的key还在
	now := time.Now().Add(delta)
	nextTime := _cronSchedule.Next(now)
	expiration := nextTime.Sub(now) + 1*time.Second // 避免在执行时过期

	ok, err := w.runScript(w.worker.ctx,
		cronOnceScript,
		[]string{key},
		now.UnixNano(),            // ARGV[1]
		nextTime.UnixNano(),       // ARGV[2]
		expiration.Milliseconds(), // ARGV[3]
	).Int()

	if err != nil { // redis报错只记录日志。
		w.worker.logger.Errorf("[CronJob]run script failed when wrapper: %v", err)
	}

	w.worker.logger.Infof("[CronJob]initialize cron once job \"%s\" for cluster, next time: %s, ok: %v", key, nextTime, ok)

	return func(ctx context.Context) {
		logger := w.worker.logger.WithContext(ctx)

		delta = w.worker.store.ServerTimeDelta(ctx)
		now = time.Now().Add(delta)
		nextTime = _cronSchedule.Next(now)
		expiration = nextTime.Sub(now) + 1*time.Second

		ok, err = w.runScript(ctx,
			cronOnceScript,
			[]string{key},
			now.UnixNano(),
			nextTime.UnixNano(),
			expiration.Milliseconds(),
		).Int()

		if err != nil { // 不能因为redis报错而跳过执行，只记录日志。
			logger.Errorf("[CronJob]run script failed: %v", err)
		} else if ok <= 0 { // 未到执行时间，跳过
			logger.Infof("[CronJob]skip the job \"%s\"", key)
			return
		}

		job(ctx)

	}
}

func (w *onceWorker) wrapperTimerJob(key string, interval time.Duration, job task.Job) task.Job {
	if key == "" {
		return job
	}
	// 先尝试加锁
	_, _ = w.runScript(w.worker.ctx, onceJobScript, []string{key}, w.worker.app.ID(), time.Duration(float64(interval)*1.1).Milliseconds()).Bool()

	return func(ctx context.Context) {
		// 判断自己是否能运行，并延长锁的过期时间
		ok, err := w.runScript(ctx, onceJobScript, []string{key}, w.worker.app.ID(), time.Duration(float64(interval)*1.1).Milliseconds()).Bool()
		if ok || err != nil {
			job(ctx)
		}
	}
}

func (w *onceWorker) WithContext(ctx context.Context) IWorker {
	return &onceWorker{
		key:    w.key,
		worker: w.worker.WithContext(ctx).(*Worker),
	}
}

// OnceForCluster 表示在后面调用的job只会在集群中执行一次。
// 如果是cron/timer任务，表示在每次定时任务触发时只在一个节点执行。
//
//	比如：OnceForCluster("key-123").Submit(func(ctx){...})
func (w *onceWorker) OnceForCluster(key string, opts ...onceOption) IWorker {
	return w.worker.OnceForCluster(key, opts...)
}

func (w *onceWorker) Submit(_job task.Job) task.TaskID {
	taskId := task.TaskID(w.key)
	_immediateSchedule := newImmediateSchedule()

	if w.GetTask(taskId) != nil && !w.override {
		return taskId
	}

	return w.worker.AddTask(taskId, _immediateSchedule, w.wrapperOnceJob(w.worker.ctx, w.key, 5*time.Second, _job))
}

func (w *onceWorker) SubmitSync(job task.JobWithError) error {
	// SubmitSync 没有taskId，不存在 override taskId
	return w.worker.SubmitSync(w.wrapperOnceJobWithError(w.worker.ctx, w.key, 5*time.Second, job))
}

func (w *onceWorker) SubmitTimer(interval time.Duration, times int64, _job task.Job) task.TaskID {
	taskId := task.TaskID(w.key)
	if w.GetTask(taskId) != nil && !w.override {
		return taskId
	}

	_timerSchedule := newTimerSchedule(interval, times)
	return w.worker.AddTaskWithCallback(
		taskId,
		_timerSchedule,
		w.wrapperTimerJob(w.key, _timerSchedule.interval, _job),
		func(taskId task.TaskID, task *task.Task) {
			_, _ = w.worker.store.Del(task.GetContext(), w.key)
		},
	)
}

func (w *onceWorker) SubmitAfter(delay time.Duration, job task.Job) task.TaskID {
	return w.SubmitTimer(delay, 1, job)
}

func (w *onceWorker) SubmitLoop(interval time.Duration, job task.Job) task.TaskID {
	return w.SubmitTimer(interval, -1, job)
}

func (w *onceWorker) Cron(spec any, _job task.Job) (task.TaskID, error) {
	// 将spec转换为cron.Schedule
	_schedule, err := w.worker.parseSchedule(spec)
	if err != nil {
		w.worker.logger.Errorf("[Cron]parse schedule %v failed: %v", spec, err)
		return "", err
	}

	taskId := task.TaskID(w.key)
	if w.GetTask(taskId) != nil && !w.override {
		return taskId, nil
	}

	_cronSchedule := newCronSchedule(_schedule)
	return w.worker.AddTask(taskId, _cronSchedule, w.wrapperOnceCronJob(w.key, _cronSchedule, _job)), nil
}

func (w *onceWorker) CronWith(job task.Job) schedule.Spec {
	// 使用当前的w.Cron，这样会执行w.wrapperOnceCronJob，保证在每次定时任务触发时只在一个节点执行
	return schedule.NewSpec(w.Cron, job)
}

func (w *onceWorker) GetTask(taskId task.TaskID) *task.Task {
	return w.worker.GetTask(taskId)
}

func (w *onceWorker) CancelTask(taskId task.TaskID) {
	w.worker.CancelTask(taskId)
}
