package task

import (
	"context"
	"github.com/robfig/cron/v3"
	"time"
)

type TaskID string

type TaskSchedule interface {
	cron.Schedule
	AddRunTimes(val int64) (newValue int64)
	CanRun() bool
	HasNext() bool
	RunTimes() int64
}

type Task struct {
	schedule TaskSchedule

	ctx context.Context
	job Job

	// createdAt is the time the task was created.
	createdAt time.Time
	// lastRunAt is the time the task was last run.
	lastRunAt  time.Time
	ScheduleId cron.EntryID
}

func NewTask(ctx context.Context, schedule TaskSchedule, job Job) *Task {
	return &Task{
		ctx:       ctx,
		createdAt: time.Now(),
		lastRunAt: time.Time{},
		job:       job,
		schedule:  schedule,
	}
}

// RunTimes 运行次数
func (t *Task) RunTimes() int64 {
	return t.schedule.RunTimes()
}

// HasNext 是否还有下一轮
func (t *Task) HasNext() bool {
	return t.schedule.HasNext()
}

// Next 下一次执行时间
func (t *Task) Next(t2 time.Time) time.Time {
	return t.schedule.Next(t2)
}

// CreatedAt 创建时间
func (t *Task) CreatedAt() time.Time {
	return t.createdAt
}

// LastRunAt 上次执行时间
func (t *Task) LastRunAt() time.Time {
	return t.lastRunAt
}

func (t *Task) GetContext() context.Context {
	return t.ctx
}

// Execute 执行Job
func (t *Task) Execute() bool {
	if !t.schedule.CanRun() {
		return false
	}
	t.lastRunAt = time.Now()
	t.schedule.AddRunTimes(1)
	t.job(t.ctx)
	return true
}
