package worker

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/schedule"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/task"
	"time"
)

type IWorker interface {
	WithContext(ctx context.Context) IWorker
	// OnceForCluster 集群中，当前任务只能执行1次
	//  - 如果是Submit，则集群中只能执行1次
	//  - 如果是Timer（涵盖SubmitAfter、SubmitLoop），同一个interval周期时，只有1个会成功执行
	//  - 如果是Cron，同一个interval周期内，只有1个会执行成功
	//  - 注意：Timer、Cron中，interval会以该key第一个提交任务的时间作为起点
	OnceForCluster(key string) IWorker
	// Submit 提交一个异步任务，运行结束后，TaskID会被删除
	//  一般情况下会立即异步执行，除非任务太多导致堆积没有被执行，才能被 CancelTask 取消
	Submit(job task.Job) task.TaskID
	// SubmitSync 提交一个同步任务，函数会等待job执行结束后返回
	SubmitSync(job task.JobWithError) error
	// SubmitTimer 提交一个定时任务,
	//  - interval: 每次运行间隔时间
	//  - times: 运行次数, -1表示无限循环，0表示不运行
	//  任务结束之后，TaskID会被删除
	SubmitTimer(interval time.Duration, times int64, job task.Job) task.TaskID
	// SubmitAfter 提交一个延迟任务，延迟时间delay后执行。运行结束后，TaskID会被删除
	//  - delay: 延迟时间
	//  任务结束之后，TaskID会被删除
	SubmitAfter(delay time.Duration, job task.Job) task.TaskID
	// SubmitLoop 提交一个无限循环任务，每次运行间隔时间interval
	//  - interval: 每次运行间隔时间
	SubmitLoop(interval time.Duration, job task.Job) task.TaskID
	// Cron 提交一个定时任务，使用cron表达式
	//  - spec: cron表达式、详情请查看github.com/robfig/cron
	Cron(spec any, job task.Job) (task.TaskID, error)
	// CronWith 提交一个定时任务，需要链式定义时间，比如：CronWith(...).EveryMinute()
	CronWith(job task.Job) schedule.Spec
	// GetTask 获取任务
	GetTask(taskId task.TaskID) *task.Task
	// CancelTask 取消任务，可以取消Submit、SubmitTimer、SubmitAfter、SubmitLoop、Cron、CronWith提交的任务
	CancelTask(jobID task.TaskID)
}
