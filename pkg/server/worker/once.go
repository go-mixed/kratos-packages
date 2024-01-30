package worker

import (
	"context"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/job"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/schedule"
	"time"
)

type onceWorker struct {
	key    string
	worker *Worker
}

var _ IWorker = (*onceWorker)(nil)

type lastRunning struct {
	// 最后一次的执行时间
	LastAt time.Time `json:"last_at" redis:"last_at"`
	// 执行的APP ID，用于区分不同的APP
	AppID string `json:"app_id" redis:"app_id"`
}

// wrapperOnceJob 保证整个集群中，该key的job只会执行一次
func (w *onceWorker) wrapperOnceJob(key string, job job.Job) job.Job {
	if key == "" {
		return job
	}
	return func(ctx context.Context) {
		// 通过setnx来保证只执行一次
		ok, err := w.worker.cache.SetNX(ctx, key, &lastRunning{
			LastAt: time.Now(),
			AppID:  w.worker.app.ID(),
		})
		if err != nil { // 不能因为redis报错而跳过执行，只记录日志。
			w.worker.logger.WithContext(ctx).Errorf("[Job]setnx %s failed: %v", key, err)
			job(ctx)
		} else if ok { // key不存在，可以执行
			job(ctx)
		} else {
			w.worker.logger.WithContext(ctx).Infof("[Job]key %s already exists, skip the job", key)
		}
	}
}

// wrapperOnceJobWithError 保证整个集群中，该key的job只会执行一次
func (w *onceWorker) wrapperOnceJobWithError(key string, job job.JobWithError) job.JobWithError {
	if key == "" {
		return job
	}
	return func(ctx context.Context) error {
		// 通过setnx来保证只执行一次
		ok, err := w.worker.cache.SetNX(ctx, key, time.Now())
		if err != nil { // 不能因为redis报错而跳过执行，只记录日志。
			w.worker.logger.WithContext(ctx).Errorf("[JobWithError]setnx %s failed: %v", key, err)
			return job(ctx)
		} else if ok { // key不存在，可以执行
			return job(ctx)
		} else {
			w.worker.logger.WithContext(ctx).Infof("[JobWithError]key %s already exists, skip the job", key)
			return nil
		}
	}
}

// onceCronRedisScript 保证整个集群中，每次cron触发时，该key的job只会执行一次。LUA可以保证原子性。
const onceCronRedisScript = `local key = KEYS[1]
local now = tonumber(ARGV[1])
local next_at = tonumber(ARGV[2])
local expiration = tonumber(ARGV[3])
local app_id = ARGV[4]
local last = redis.call('get', key)
local res = 0
if last then
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
else -- key不存在，可以执行
	res = 1
end
-- res > 0 表示可以执行，设置key和下次执行时间
redis.call('set', key, cjson.encode({last_at = now, next_at = next_at, app_id = app_id}), 'px', expiration)
return res
`

// wrapperOnceCronJob 保证整个集群中，每次cron触发时，该key的job只会执行一次
// （注意：基于的是每次cron触发时的时间点，比如任务是EveryMinute，那么表示每分钟在集群中只会执行一次）
func (w *onceWorker) wrapperOnceCronJob(key string, spec cron.Schedule, job job.Job) job.Job {
	if key == "" || spec == nil {
		return job
	}

	// 注册脚本
	script := w.worker.cache.Script(onceCronRedisScript)
	// 为了确保cron的多个节点的时间一致，这里计算出redis服务器时间与本地时间的差值，
	// 后面的now, nextTime都根据delta修正为redis服务器时间
	delta := w.worker.cache.ServerTimeDelta(context.Background())

	// 录入cron任务时，如果key不存在，就设置下次运行的时间
	// 比如：程序滚动发布时，上一个执行的key还在
	now := time.Now().Add(delta)
	nextTime := spec.Next(now)
	expiration := nextTime.Sub(now) * 2 // 双倍过期时间，避免在执行时过期

	ok, err := script.Run(context.Background(),
		[]string{key},
		now.UnixNano(),            // ARGV[1]
		nextTime.UnixNano(),       // ARGV[2]
		expiration.Milliseconds(), // ARGV[3]
		w.worker.app.ID()).Int()   // ARGV[4]

	if err != nil { // redis报错只记录日志。
		w.worker.logger.Errorf("[CronJob]run script failed when wrapper: %v", err)
	}

	w.worker.logger.Infof("[CronJob]initialize cron once job \"%s\" for cluster, next time: %s, ok: %v", key, nextTime, ok)

	return func(ctx context.Context) {
		logger := w.worker.logger.WithContext(ctx)

		delta = w.worker.cache.ServerTimeDelta(ctx)
		now = time.Now().Add(delta)
		nextTime = spec.Next(now)
		expiration = nextTime.Sub(now) * 2

		ok, err = script.Run(ctx,
			[]string{key},
			now.UnixNano(),
			nextTime.UnixNano(),
			expiration.Milliseconds(),
			w.worker.app.ID()).Int()

		if err != nil { // 不能因为redis报错而跳过执行，只记录日志。
			logger.Errorf("[CronJob]run script failed: %v", err)
		} else if ok <= 0 { // 未到执行时间，跳过
			logger.Infof("[CronJob]skip the job \"%s\"", key)
			return
		}

		job(ctx)

	}
}

func (w *onceWorker) WithContext(ctx context.Context) IWorker {
	return &onceWorker{
		key:    w.key,
		worker: w.worker.WithContext(ctx).(*Worker),
	}
}

// OnceForCluster 表示在后面调用的job只会在集群中执行一次。
// 如果是cron任务，表示在每次定时任务触发时只在一个节点执行。
//
//	比如：OnceForCluster("key-123").Submit(func(ctx){...})
func (w *onceWorker) OnceForCluster(key string, options ...onceOption) IWorker {
	return w.worker.OnceForCluster(key, options...)
}

func (w *onceWorker) Submit(job job.Job) {
	w.worker.Submit(w.wrapperOnceJob(w.key, job))
}

func (w *onceWorker) SubmitWait(job job.Job) {
	w.worker.SubmitWait(w.wrapperOnceJob(w.key, job))
}

func (w *onceWorker) SubmitAfter(delay time.Duration, job job.Job) {
	w.worker.SubmitAfter(delay, w.wrapperOnceJob(w.key, job))
}

func (w *onceWorker) SubmitWithError(job job.JobWithError) error {
	return w.worker.SubmitWithError(w.wrapperOnceJobWithError(w.key, job))
}

func (w *onceWorker) Cron(spec any, job job.Job) (cron.EntryID, error) {
	// 将spec转换为cron.Schedule
	_schedule, err := w.worker.parseSchedule(spec)
	if err != nil {
		w.worker.logger.Errorf("[Cron]parse schedule %v failed: %v", spec, err)
		return 0, err
	}

	return w.worker.Cron(_schedule, w.wrapperOnceCronJob(w.key, _schedule, job))
}

func (w *onceWorker) CronWith(job job.Job) schedule.Spec {
	// 使用当前的w.Cron，这样会执行w.wrapperOnceCronJob，保证在每次定时任务触发时只在一个节点执行
	return schedule.NewSpec(w.Cron, job)
}
