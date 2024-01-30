package worker

import (
	"context"
	"fmt"
	"github.com/RussellLuo/timingwheel"
	"github.com/gammazero/workerpool"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/job"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/schedule"
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
	app    *app.App
	logger *log.Helper
	cache  *cache.Cache

	pool           *workerpool.WorkerPool
	timeWheel      *timingwheel.TimingWheel
	schedule       *cron.Cron
	scheduleParser cron.Parser
	stopped        *atomic.Bool
	ctx            context.Context
}

var _ transport.Server = (*Worker)(nil)
var _ IWorker = (*Worker)(nil)

func NewWorker(
	app *app.App,
	logger log.Logger,
	cache *cache.Cache,

	maxWorkers int,
) *Worker {
	scheduleParser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	return &Worker{
		app:    app,
		logger: log.NewModuleHelper(logger, "worker"),
		cache:  cache,

		pool:      workerpool.New(maxWorkers),
		timeWheel: timingwheel.NewTimingWheel(time.Millisecond, 20),
		schedule: cron.New(
			cron.WithParser(scheduleParser),
			cron.WithLogger(schedule.NewScheduleLogger(logger)),
		),
		scheduleParser: scheduleParser,
		stopped:        &atomic.Bool{},
		ctx:            app.BaseContext(),
	}
}

// clone returns a new worker with the same configuration, which will not affect the execution of the task.
// clone 返回一个拥有相同的配置的新的worker，不会影响任务的执行
func (w *Worker) clone() *Worker {
	return &Worker{
		app:    w.app,
		logger: w.logger,
		cache:  w.cache.Clone(),

		pool:      w.pool,
		timeWheel: w.timeWheel,
		schedule:  w.schedule,
		stopped:   w.stopped,
		ctx:       w.ctx,
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
// execute only once in the cluster. If it is a cron task, it means that only one node is executed at a time.
// e.g.: OnceForCluster("key-123").Submit(func(ctx){...}) means that this key-123 job will only be executed once in the cluster.
// Warning: The program can only guarantee that the job will only be executed once within the cache.expiration time.
// If the key expires, the subsequent submitted jobs with the same name will still be executed.
// If it is a cron task, it means that only one node is executed at a time. (Not controlled by cache.expiration)
//
//	OnceForCluster 表示该key的job只会在集群中执行一次。
//	比如：OnceForCluster("key-123").Submit(func(ctx){...})，表示这个key-123的job只会在集群中执行一次。
//	注意：程序只能保证在cache.expiration时间内，只会执行一次，如果key过期了，后续提交后的同名job还是会执行。
//	如果是cron任务，表示在每次定时任务触发时只在一个节点执行。（不受cache.expiration过期控制）
func (w *Worker) OnceForCluster(key string, options ...onceOption) IWorker {
	ow := &onceWorker{
		key:    key,
		worker: w.clone(),
	}

	for _, option := range options {
		option(ow)
	}

	return ow
}

// Submit submits a task to be executed by a worker.
// It's a NON-BLOCKING call, aka an ASYNCHRONOUS task.
// (The job will use the converted context without canceling when calling)
//
//	Submit 提交一个任务给worker执行，这是一个异步调用；
//	（为了防止job在调用时会用到request canceled的ctx，将会context转换成一个不会cancel的context）
func (w *Worker) Submit(job job.Job) {
	ctx := w.app.CloneContextFromBase(w.ctx)
	w.pool.Submit(func() {
		job(ctx)
	})
}

// SubmitWait submits a task to be executed by a worker.
// It's a blocking call until the worker finishes the task
// (The job will use the app.BaseContext() when calling， or you defined the context by WithContext)
//
//	SubmitWait 提交一个任务给worker执行，这是一个同步调用，必须等待任务执行完毕才会返回。
//	（job执行时使用的是app.BaseContext()，或者你可以通过WithContext自定义context）
func (w *Worker) SubmitWait(job job.Job) {
	// ctx := w.app.CloneContextFromBase(w.ctx)
	w.pool.SubmitWait(func() {
		job(w.ctx)
	})
}

// SubmitWithError submits a task to be executed by a worker and returns the error.
// It's a blocking call until the worker finishes the task
// (The job will use the app.BaseContext() when calling， or you defined the context by WithContext)
//
//	SubmitWithError 提交一个任务给worker执行，这是一个同步调用，必须等待任务执行完毕才会返回error。
//	（job执行时使用的是app.BaseContext()，或者你可以通过WithContext自定义context）
func (w *Worker) SubmitWithError(job job.JobWithError) error {
	// ctx := w.app.CloneContextFromBase(w.ctx)
	var err error
	w.pool.SubmitWait(func() {
		err = job(w.ctx)
	})

	return err
}

// SubmitAfter submits a task to be executed by a worker after a delay.
// It's a NON-BLOCKING call, aka an ASYNCHRONOUS task.
// (The job will use the converted context without canceling when calling)
//
//	SubmitAfter 提交一个任务在delay之后执行，这是一个【异步】调用；
//	（为了防止job在调用时会用到request canceled的ctx，将会context转换成一个不会cancel的context）
func (w *Worker) SubmitAfter(delay time.Duration, job job.Job) {
	ctx := w.app.CloneContextFromBase(w.ctx)
	w.timeWheel.AfterFunc(delay, func() {
		w.pool.Submit(func() {
			job(ctx)
		})
	})
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
// It's a NON-BLOCKING call, aka an ASYNCHRONOUS task.
// (The job will use the app.BaseContext() when calling， or you defined the context by WithContext)
//
//	Cron 添加一个cron任务给worker执行，这是一个【异步】调用；
//	（job执行时使用的是app.BaseContext()，或者你可以通过WithContext自定义context）
//	支持的表达式： https://pkg.go.dev/github.com/robfig/cron/v3#hdr-Special_Characters
func (w *Worker) Cron(spec any, job job.Job) (cron.EntryID, error) {
	expr, err := w.parseSchedule(spec)
	if err != nil {
		return 0, err
	}

	return w.schedule.Schedule(expr, cron.FuncJob(func() {
		job(w.ctx)
	})), nil
}

// CronWith add a cron job to the worker with a chain caller: w.CronWith(func(ctx){...}).Every(30 * time.Second)
// It's a NON-BLOCKING call, aka an ASYNCHRONOUS task.
// (The job will use the app.BaseContext() when calling， or you defined the context by WithContext)
//
//	CronWith 添加一个cron的任务给worker执行（链式调用），这是一个【异步】调用：w.CronWith(job).Every(30 * time.Second)
//	（job执行时使用的是app.BaseContext()，或者你可以通过WithContext自定义context）
func (w *Worker) CronWith(job job.Job) schedule.Spec {
	return schedule.NewSpec(w.Cron, job)
}

func (w *Worker) Start(ctx context.Context) error {
	w.ctx = ctx
	w.stopped.Store(false)

	w.timeWheel.Start()
	w.schedule.Start()
	w.logger.WithContext(ctx).Infof("time wheel, schedule, worker pool(size=%d) started", w.pool.Size())
	return nil
}

func (w *Worker) Stop(ctx context.Context) error {
	w.ctx = ctx
	w.pool.StopWait()
	w.timeWheel.Stop()
	w.schedule.Stop()
	w.stopped.Store(true)

	w.logger.WithContext(ctx).Infof("time wheel, schedule, worker pool server stop")
	return nil
}

func (w *Worker) Stopped() bool {
	return w.stopped.Load()
}
