package worker

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/job"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/schedule"
	"time"
)

type onceWorker struct {
	key string
	// worker 外部创建时，必须是worker的clone体。因为会修改worker.store的属性
	worker *Worker
}

var _ IWorker = (*onceWorker)(nil)

func (w *onceWorker) runScript(ctx context.Context, code string, keys []string, args ...any) *redis.Cmd {
	script := w.worker.store.Script(code)
	return script.Run(ctx, keys, args...)
}

const createOnceJobScript = `local key = KEYS[1]
local app_id = ARGV[1]
local uuid = ARGV[2]
local expiration = tonumber(ARGV[3]) or 0 -- 如果 ARGV[3] 不是数字，则设置为 0

-- 序列化表
local str = cjson.encode({
    app_id = app_id,
    uuid = uuid,
    last_at = redis.call('TIME'),
})

-- 尝试设置键值对
local res = redis.call('SETNX', key, str)

-- 如果res为1，表示键是新设置的，设置过期时间
if res == 1 then
    redis.call('PEXPIRE', key, expiration)
	return true
end

return false -- 返回SETNX的结果，true表示新设置，false表示已存在
`

const renewalOnceJobScript = `local key = KEYS[1]
local app_id = ARGV[1]
local uuid = ARGV[2]
local expiration = tonumber(ARGV[3]) or 0 -- 如果 ARGV[3] 不是数字，则设置为 0

local existing_str = redis.call('get', key)
if not existing_str then
	return false -- key不存在，无法续期
end

local existing_obj, err = cjson.decode(existing_str)
if existing_obj and type(existing_obj) == 'table' and existing_obj['uuid'] == uuid and existing_obj['app_id'] == app_id then -- 检查existing_obj是否为table，以及uuid和app_id是否匹配
	-- 如果匹配，续期键的过期时间
	redis.call('PEXPIRE', key, expiration)
	return true -- 表示续期成功
end

return false -- 表示续期失败或不匹配
`

// 删除指定UUID的job
const deleteOnceJobScript = `local key = KEYS[1]
local app_id = ARGV[1]
local uuid = ARGV[2]
local str = redis.call('get', key)
if not str then
    return false
end

local res = false

local obj, err = cjson.decode(str)
if err or obj == nil or type(obj) ~= 'table' then -- 如果解码失败，可以删除
    res = true
elseif obj['uuid'] == uuid and obj['app_id'] == app_id then -- 如果值匹配，可以删除
	res = true
end

if res then
    -- 如果条件满足，删除键
    redis.call('del', key)
end

return res
`

// executeOnceJob 执行1次job：保证在集群中，这个job执行期间内，同名key绝对只会执行1次
func (w *onceWorker) executeOnceJob(ctx context.Context, key string, job job.JobWithError) error {
	logger := w.worker.logger.WithContext(ctx)
	appId := w.worker.app.ID()
	uuid := uuid.New().String()
	expiration := time.Second * 5

	// 通过lua来创建key，并设置过期时间，确保本job在当前时间段内只会被执行1次
	ok, err := w.runScript(ctx, createOnceJobScript, []string{key}, appId, uuid, expiration).Bool()
	if err != nil { // 不能因为redis报错而跳过执行，只记录日志。
		logger.Errorf("[JobWithError]setnx %s failed: %v", key, err)
		return job(ctx)
	} else if !ok {
		logger.Infof("[JobWithError]key %s already exists, skip the job", key)
		return nil
	} // key不存在，可以执行

	// 创建一个ticker，用于检查key是否过期
	ticker := w.worker.WithContext(ctx).SubmitTicker(expiration/2, func(ctx context.Context) {
		if err1 := w.runScript(ctx, renewalOnceJobScript, []string{key}, appId, uuid, expiration).Err(); err1 != nil {
			logger.Errorf("[JobWithError]renewal key %s failed: %v", key, err1)
		}
	})

	// 任务执行完毕，停止ticker，删除key
	defer func() {
		if rErr := recover(); rErr != nil {
			logger.Errorf("[Job]panic: %v", rErr)
		}
		ticker.Stop()
		if err = w.runScript(ctx, deleteOnceJobScript, []string{key}, appId, uuid).Err(); err != nil {
			logger.Errorf("[JobWithError]delete key %s failed: %v", key, err)
		}
	}()

	return job(ctx)
}

// wrapperOnceJob 保证整个集群中，该key的job只会执行一次
//
//	控制的是整个job的执行生命周期。job执行完毕后，才能第二次执行
func (w *onceWorker) wrapperOnceJob(key string, job job.Job) job.Job {
	if key == "" {
		return job
	}

	return func(ctx context.Context) {
		_ = w.executeOnceJob(ctx, key, func(ctx context.Context) error {
			job(ctx)
			return nil
		})
	}
}

// wrapperOnceJobWithError 保证整个集群中，该key的job只会执行一次
func (w *onceWorker) wrapperOnceJobWithError(key string, job job.JobWithError) job.JobWithError {
	if key == "" {
		return job
	}
	return func(ctx context.Context) error {
		return w.executeOnceJob(ctx, key, job)
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
	script := w.worker.store.Script(cronOnceScript)
	// 为了确保cron的多个节点的时间一致，这里计算出redis服务器时间与本地时间的差值，
	// 后面的now, nextTime都根据delta修正为redis服务器时间
	delta := w.worker.store.ServerTimeDelta(context.Background())

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

		delta = w.worker.store.ServerTimeDelta(ctx)
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
func (w *onceWorker) OnceForCluster(key string) IWorker {
	return w.worker.OnceForCluster(key)
}

func (w *onceWorker) Submit(job job.Job) {
	w.worker.Submit(w.wrapperOnceJob(w.key, job))
}

func (w *onceWorker) SubmitWait(job job.Job) {
	w.worker.SubmitWait(w.wrapperOnceJob(w.key, job))
}

func (w *onceWorker) SubmitAfter(delay time.Duration, job job.Job) *time.Timer {
	return w.worker.SubmitAfter(delay, w.wrapperOnceJob(w.key, job))
}

func (w *onceWorker) SubmitTicker(interval time.Duration, job job.Job) *time.Ticker {
	return w.worker.SubmitTicker(interval, w.wrapperOnceJob(w.key, job))
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
