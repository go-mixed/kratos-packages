package worker

import (
	"context"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/schedule"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/task"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"sync/atomic"
	"time"
)

// Worker creates a new worker server.
// It's used for processing some scattered tasks, such as sending messages, emails, delayed tasks, etc.
// The current task pool will limit the number of concurrent tasks,
// It will reduce the overhead of context switching compared to directly creating multiple goroutines.
//
//	异步任务服务，用于处理一些零散的任务，比如发送消息、发送邮件、延迟任务等。
//	会限制并发数，但是相比直接新建多个go协程，这个会明显多个协程导致的频繁上下文切换的开销。
type Worker struct {
	app     *app.App
	logger  *log.Helper
	store   *redis.Redis
	options workerOptions

	pool           *workerpool.WorkerPool
	schedule       *cron.Cron
	scheduleParser cron.Parser
	stopped        *atomic.Bool
	ctx            context.Context

	// 记录当前正在执行的任务
	tasks utils.ConcurrentMap[task.TaskID, *task.Task]
}

var _ transport.Server = (*Worker)(nil)
var _ IWorker = (*Worker)(nil)

func NewWorker(
	app *app.App,
	logger log.Logger,
	rdb *redis.Client,

	options ...workerOption,
) *Worker {

	var opts workerOptions = DefaultWorkerOptions()
	for _, option := range options {
		option(&opts)
	}

	store := redis.NewRedis(rdb.WithTimeout(rdb.Options().ReadTimeout), opts.redisOptions)
	scheduleParser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	return &Worker{
		app:     app,
		logger:  log.NewModuleHelper(logger, "worker"),
		store:   store,
		options: opts,

		pool: workerpool.New(opts.workerCount),
		schedule: cron.New(
			cron.WithParser(scheduleParser),
			cron.WithLogger(schedule.NewScheduleLogger(logger)),
		),
		scheduleParser: scheduleParser,
		stopped:        &atomic.Bool{},
		ctx:            app.BaseContext(),

		tasks: utils.ConcurrentMap[task.TaskID, *task.Task]{},
	}
}

// clone returns a new worker with the same configuration, which will not affect the execution of the task.
// clone 返回一个拥有相同的配置的新的worker，不会影响任务的执行
func (w *Worker) clone() *Worker {
	return &Worker{
		app:    w.app,
		logger: w.logger,
		store:  w.store.Clone(),

		pool:           w.pool,
		schedule:       w.schedule,
		scheduleParser: w.scheduleParser,
		stopped:        w.stopped,
		ctx:            w.ctx,
	}
}

// WithContext returns a new worker with the given context.
//
//	WithContext 使用给定的context返回一个新的worker。
func (w *Worker) WithContext(ctx context.Context) IWorker {
	_w := w.clone()
	_w.ctx = ctx
	return _w
}

// OnceForCluster submits a task to be executed by a worker.
// execute only once in the cluster. If it is a cron/timer task, it means that only one node is executed at a time.
//
//	OnceForCluster 表示该key的job只会在集群中执行一次。
//	比如：OnceForCluster("key-123").Submit(func(ctx){...})，表示这个key-123的job只会在集群中执行一次。
//	如果是cron/timer任务，表示在每次定时任务触发时只在一个节点执行。
func (w *Worker) OnceForCluster(key string, opts ...onceOption) IWorker {
	ow := &onceWorker{
		key:         key,
		worker:      w.clone(),
		keyAsTaskID: false,
	}

	for _, opt := range opts {
		opt(ow)
	}

	return ow
}

// Submit submits a ASYNCHRONOUS task to be executed by a worker.
//
//	Submit 提交一个异步任务给worker执行。
func (w *Worker) Submit(_job task.Job) task.TaskID {
	taskId := task.TaskID(fmt.Sprintf("immediate:%s", uuid.NewString()))

	return w.AddTask(taskId, newImmediateSchedule(), _job)
}

// SubmitSync submits a task to be executed by a worker and returns the error.
// It's a blocking call until the worker finishes the task
//
//	SubmitSync 提交一个任务给worker执行，这是一个同步调用，必须等待任务执行完毕才会返回error。
func (w *Worker) SubmitSync(job task.JobWithError) error {
	var err error

	w.pool.SubmitWait(func() {
		err = job(w.ctx)
	})
	return err
}

// AddTask adds a ASYNCHRONOUS task to be executed by a worker.
//
//	AddTask 添加一个任务给worker执行，这是一个异步调用；
func (w *Worker) AddTask(
	taskId task.TaskID,
	schedule task.TaskSchedule,
	_job task.Job,
) task.TaskID {
	return w.AddTaskWithCallback(taskId, schedule, _job, nil)
}

// AddTaskWithCallback adds a ASYNCHRONOUS task to be executed by a worker. and callback when task complete
//
//	AddTaskWithCallback 添加一个异步任务给worker执行，在执行完毕之后，调用onComplete；
func (w *Worker) AddTaskWithCallback(
	taskId task.TaskID,
	schedule task.TaskSchedule,
	_job task.Job,
	onComplete func(taskId task.TaskID, task *task.Task),
) task.TaskID {
	_task, ok := w.tasks.Load(taskId)
	// 存在旧的任务，移除旧的任务
	if ok && _task.ScheduleId != 0 {
		w.CancelTask(taskId)
	}

	// 创建task，以及添加到cron任务列表中
	ctx := w.app.CloneContextFromBase(w.ctx)
	task := task.NewTask(ctx, schedule, _job)

	// 是立即执行的任务，调用pool直接执行
	if IsImmediateSchedule(schedule) {
		w.tasks.Store(taskId, task)

		w.pool.Submit(func() {
			_task, ok := w.tasks.Load(taskId)
			if !ok { // 不存在Key，说明task已经被cancel了
				return
			}

			// 执行完毕之后，移除task
			_task.Execute()
			w.CancelTask(taskId)
		})
		return taskId
	}

	// delay task，丢入Cron中
	task.ScheduleId = w.schedule.Schedule(task, cron.FuncJob(func() {
		_task, ok := w.tasks.Load(taskId)
		if !ok { // 不存在Key，说明task属于悬挂状态
			w.schedule.Remove(_task.ScheduleId)
			return
		}
		// 执行，并返回是否执行
		_task.Execute()

		// 如果没有下一次执行，移除任务，并回调onComplete
		if !_task.HasNext() {
			w.CancelTask(taskId) // 移除任务
			if onComplete != nil {
				onComplete(taskId, _task)
			}
		}
	}))

	w.tasks.Store(taskId, task)
	return taskId
}

// SubmitTimer submits a task to be executed by a worker every interval until it reaches times.
func (w *Worker) SubmitTimer(interval time.Duration, times int64, _job task.Job) task.TaskID {

	taskId := task.TaskID(fmt.Sprintf("timer:%s", uuid.NewString()))
	timerSchedule := newTimerSchedule(interval, times)
	return w.AddTask(taskId, timerSchedule, _job)
}

// SubmitAfter submits a ASYNCHRONOUS task to be executed by a worker after a delay.
//
//	SubmitAfter 提交一个异步任务在delay之后执行。
func (w *Worker) SubmitAfter(delay time.Duration, job task.Job) task.TaskID {
	return w.SubmitTimer(delay, 1, job)
}

// SubmitLoop submits a ASYNCHRONOUS task to be executed by a worker every interval.
//
//	SubmitLoop 提交一个异步任务在每个interval时执行；
func (w *Worker) SubmitLoop(interval time.Duration, job task.Job) task.TaskID {
	return w.SubmitTimer(interval, -1, job)
}

func (w *Worker) parseSchedule(spec any) (cron.Schedule, error) {
	switch expr := spec.(type) {
	case cron.Schedule:
		return expr, nil
	case string:
		return w.scheduleParser.Parse(expr)
	default:
		return nil, fmt.Errorf("不支持此类型的解析表达式, %v", spec)
	}
}

// Cron add a cron job to the worker.
//
//	Cron 添加一个cron任务给worker执行，这是一个【异步】调用；
//	支持的表达式： https://pkg.go.dev/github.com/robfig/cron/v3#hdr-Special_Characters
func (w *Worker) Cron(spec any, _job task.Job) (task.TaskID, error) {
	expr, err := w.parseSchedule(spec)
	if err != nil {
		return "", err
	}

	cronSchedule := newCronSchedule(expr)
	taskId := task.TaskID(fmt.Sprintf("cron:%d", uuid.NewString()))
	return w.AddTask(taskId, cronSchedule, _job), nil
}

// CronWith add a cron job to the worker with a chain caller: w.CronWith(func(ctx){...}).Every(30 * time.Second)
//
//	CronWith 添加一个cron的任务给worker执行（链式调用）：w.CronWith(job).Every(30 * time.Second)
func (w *Worker) CronWith(job task.Job) schedule.Spec {
	return schedule.NewSpec(w.Cron, job)
}

// GetTask get a task by taskId
func (w *Worker) GetTask(taskId task.TaskID) *task.Task {
	task, ok := w.tasks.Load(taskId)
	if !ok {
		return nil
	}
	return task
}

// CancelTask cancel a task by taskId
func (w *Worker) CancelTask(jobID task.TaskID) {
	task, ok := w.tasks.LoadAndDelete(jobID)
	if !ok {
		return
	} else if task.ScheduleId != 0 { // remove cron task
		w.schedule.Remove(task.ScheduleId)
	}
}

func (w *Worker) Start(ctx context.Context) error {
	w.ctx = ctx
	w.stopped.Store(false)

	w.schedule.Start()
	w.logger.WithContext(ctx).Infof("time wheel, schedule, worker pool(size=%d) started", w.pool.Size())
	return nil
}

func (w *Worker) Stop(ctx context.Context) error {
	w.ctx = ctx
	w.pool.StopWait()
	w.schedule.Stop()
	w.stopped.Store(true)

	w.logger.WithContext(ctx).Infof("time wheel, schedule, worker pool server stop")
	return nil
}

func (w *Worker) Stopped() bool {
	return w.stopped.Load()
}
