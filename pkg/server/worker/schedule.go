package worker

import (
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/task"
	"sync/atomic"
	"time"
)

type timerSchedule struct {
	// totalTimes is the number of times the job should be run.
	totalTimes int64
	// runTimes
	runTimes atomic.Int64
	// interval is the interval between runs.
	interval time.Duration
}

var _ task.TaskSchedule = (*timerSchedule)(nil)

// newTimerSchedule returns a schedule that runs after a given interval.
func newTimerSchedule(interval time.Duration, times int64) *timerSchedule {
	return &timerSchedule{
		interval:   interval,
		totalTimes: times,
		runTimes:   atomic.Int64{},
	}
}

func IsTimerSchedule(schedule task.TaskSchedule) bool {
	_, ok := schedule.(*timerSchedule)
	return ok
}

func (t *timerSchedule) AddRunTimes(val int64) (newValue int64) {
	return t.runTimes.Add(val)
}

func (t *timerSchedule) RunTimes() int64 {
	return t.runTimes.Load()
}

func (t *timerSchedule) CanRun() bool {
	return t.totalTimes == -1 || t.runTimes.Load() < t.totalTimes
}

func (t *timerSchedule) HasNext() bool {
	return t.totalTimes == -1 || t.runTimes.Load()+1 < t.totalTimes
}

func (t *timerSchedule) Next(_time time.Time) time.Time {
	return _time.Add(t.interval)
}

type cronSchedule struct {
	schedule cron.Schedule
	runTimes atomic.Int64
}

var _ task.TaskSchedule = (*cronSchedule)(nil)

// newCronSchedule returns a schedule that runs according to the given cron schedule.
func newCronSchedule(nativeCronSchedule cron.Schedule) *cronSchedule {
	return &cronSchedule{
		schedule: nativeCronSchedule,
		runTimes: atomic.Int64{},
	}
}

func (s *cronSchedule) AddRunTimes(val int64) int64 {
	return s.runTimes.Add(val)
}

func (s *cronSchedule) RunTimes() int64 {
	return s.runTimes.Load()
}

func (s *cronSchedule) CanRun() bool {
	return true
}

func (s *cronSchedule) HasNext() bool {
	return true
}

func (s *cronSchedule) Next(_time time.Time) time.Time {
	return s.schedule.Next(_time)
}

type immediateSchedule struct {
	runTimes atomic.Int64
}

var _ task.TaskSchedule = (*immediateSchedule)(nil)

// newImmediateSchedule returns a schedule that runs immediately.
func newImmediateSchedule() *immediateSchedule {
	return &immediateSchedule{
		runTimes: atomic.Int64{},
	}
}

// IsImmediateSchedule returns true if the given schedule is an immediate schedule.
func IsImmediateSchedule(schedule task.TaskSchedule) bool {
	_, ok := schedule.(*immediateSchedule)
	return ok
}

func (s *immediateSchedule) Next(t time.Time) time.Time {
	return t
}

func (s *immediateSchedule) AddRunTimes(val int64) (newValue int64) {
	return s.runTimes.Add(val)
}

func (s *immediateSchedule) CanRun() bool {
	return s.runTimes.Load() == 0
}

func (s *immediateSchedule) HasNext() bool {
	return false
}

func (s *immediateSchedule) RunTimes() int64 {
	return s.runTimes.Load()
}
