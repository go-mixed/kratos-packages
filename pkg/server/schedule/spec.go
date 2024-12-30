package schedule

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/task"
	"math"
	"strings"
	"time"
)

type Spec interface {
	Cron(expression any) (task.TaskID, error)
	Every(duration time.Duration) (task.TaskID, error)
	EverySeconds(seconds ...int) (task.TaskID, error)
	EveryMinute() (task.TaskID, error)
	AfterEveryMinute() (task.TaskID, error)
	EveryMinutes(minutes int) (task.TaskID, error)
	AfterEveryMinutes(minutes int) (task.TaskID, error)
	EveryFiveMinutes() (task.TaskID, error)
	AfterEveryFiveMinutes() (task.TaskID, error)
	EveryTenMinutes() (task.TaskID, error)
	AfterEveryTenMinutes() (task.TaskID, error)
	EveryFifteenMinutes() (task.TaskID, error)
	AfterEveryFifteenMinutes() (task.TaskID, error)
	EveryThirtyMinutes() (task.TaskID, error)
	AfterEveryThirtyMinutes() (task.TaskID, error)
	Hourly() (task.TaskID, error)
	HourlyAt(offset int) (task.TaskID, error)
	Daily() (task.TaskID, error)
	DailyAt(t string) (task.TaskID, error)
	Weekly() (task.TaskID, error)
	Monthly() (task.TaskID, error)
}

const (
	// Set the top bit if a star was included in the expression.
	starBit = 1 << 63
)

type bounds struct {
	min, max uint
	names    map[string]uint
}

// getBits sets all bits in the range [min, max], modulo the given step size.
func getBits(min, max, step uint) uint64 {
	var bits uint64

	// If step is 1, use shifts.
	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	// Else, use a simple loop.
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

// all returns all bits within the given bounds.  (plus the star bit)
func all(r bounds) uint64 {
	return getBits(r.min, r.max, 1) | starBit
}

var _ Spec = new(spec)

type cronCaller func(spec any, job task.Job) (task.TaskID, error)

type spec struct {
	job    task.Job
	runner cronCaller
}

// NewSpec 实例化cron表达式
func NewSpec(runner cronCaller, job task.Job) *spec {
	return &spec{
		job:    job,
		runner: runner,
	}
}

// Cron 自定义cron表达式运行job
// 支持表达式：https://pkg.go.dev/github.com/robfig/cron/v3#hdr-Special_Characters
func (s spec) Cron(spec any) (task.TaskID, error) {
	return s.runner(spec, s.job)
}

// Every 传入time.Duration运行job，精度为秒
func (s spec) Every(duration time.Duration) (task.TaskID, error) {
	return s.runner(cron.Every(duration), s.job)
}

// EverySeconds 每秒运行job
func (s spec) EverySeconds(seconds ...int) (task.TaskID, error) {
	defaultSeconds := 1
	if len(seconds) > 0 {
		defaultSeconds = seconds[0]
	}
	return s.Cron(fmt.Sprintf("@every %ds", defaultSeconds))
}

// EveryMinute 每分钟运行job，会在**:**:00的每分钟运行（即秒为00）
func (s spec) EveryMinute() (task.TaskID, error) {
	return s.EveryMinutes(1)
}

// AfterEveryMinute 每分钟后运行job，设置时间之后每1分钟，比如当前时间为**00::23，则在**:**:23运行
func (s spec) AfterEveryMinute() (task.TaskID, error) {
	return s.Cron("@every 1m")
}

// EveryMinutes 每多少分钟运行job，比如5分钟，则会在这些时间运行：**:00:00, **:05:00, **:10:00, **:15:00, **:20:00, **:25:00, **:30:00, **:35:00, **:40:00, **:45:00, **:50:00, **:55:00
func (s spec) EveryMinutes(minutes int) (task.TaskID, error) {
	return s.Cron(fmt.Sprintf("0 */%d * * * *", minutes))
}

// AfterEveryMinutes 每多少分钟后运行job，比如当前时间为**00::23，设置为5，则会在这些时间运行：**:00:23, **:05:23, **:10:23, **:15:23, **:20:23, **:25:23, **:30:23, **:35:23, **:40:23, **:45:23, **:50:23, **:55:23
func (s spec) AfterEveryMinutes(minutes int) (task.TaskID, error) {
	return s.Cron(fmt.Sprintf("@every %dm", minutes))
}

// EveryFiveMinutes 每五分钟运行job，会在这些时间运行：**:00:00, **:05:00, **:10:00, **:15:00, **:20:00, **:25:00, **:30:00, **:35:00, **:40:00, **:45:00, **:50:00, **:55:00
func (s spec) EveryFiveMinutes() (task.TaskID, error) {
	return s.EveryMinutes(5)
}

// AfterEveryFiveMinutes 每五分钟后运行job，比如当前时间为**00::23，设置为5，则会在这些时间运行：**:00:23, **:05:23, **:10:23, **:15:23, **:20:23, **:25:23, **:30:23, **:35:23, **:40:23, **:45:23, **:50:23, **:55:23
func (s spec) AfterEveryFiveMinutes() (task.TaskID, error) {
	return s.AfterEveryMinutes(5)
}

// EveryTenMinutes 每十分钟运行job，在这些时间运行：**:00:00, **:10:00, **:20:00, **:30:00, **:40:00, **:50:00
func (s spec) EveryTenMinutes() (task.TaskID, error) {
	return s.EveryMinutes(10)
}

// AfterEveryTenMinutes 每十分钟后运行job，比如当前时间为**00::23，则会在这些时间运行：**:00:23, **:10:23, **:20:23, **:30:23, **:40:23, **:50:23
func (s spec) AfterEveryTenMinutes() (task.TaskID, error) {
	return s.AfterEveryMinutes(10)
}

// EveryFifteenMinutes 每十五分钟运行job，在这些时间运行：**:00:00, **:15:00, **:30:00, **:45:00
func (s spec) EveryFifteenMinutes() (task.TaskID, error) {
	return s.EveryMinutes(15)
}

// AfterEveryFifteenMinutes 每十五分钟后运行job，比如当前时间为**00::23，则会在这些时间运行：**:00:23, **:15:23, **:30:23, **:45:23
func (s spec) AfterEveryFifteenMinutes() (task.TaskID, error) {
	return s.AfterEveryMinutes(15)
}

// EveryThirtyMinutes 每三十分钟运行job，在这些时间运行：**:00:00, **:30:00
func (s spec) EveryThirtyMinutes() (task.TaskID, error) {
	return s.EveryMinutes(30)
}

// AfterEveryThirtyMinutes 每三十分钟后运行job，比如当前时间为**00::23，则会在这些时间运行：**:00:23, **:30:23
func (s spec) AfterEveryThirtyMinutes() (task.TaskID, error) {
	return s.AfterEveryMinutes(30)
}

// Hourly 每小时运行job，在**:00:00运行
func (s spec) Hourly() (task.TaskID, error) {
	return s.Cron("@hourly")
}

// HourlyAt 每小时的某分钟运行job
func (s spec) HourlyAt(offset int) (task.TaskID, error) {
	return s.Cron(fmt.Sprintf("@every 1h%dm", offset))
}

// Daily 每天运行job，在00:00:00运行
func (s spec) Daily() (task.TaskID, error) {
	return s.Cron("@daily")
}

// DailyAt 每天某时某分运行job
// DailyAt("12:21") 每天12点21分钟运行
func (s spec) DailyAt(t string) (task.TaskID, error) {
	tt := strings.Split(t, ":")
	if len(tt) < 2 {
		tt = append(tt, "00")
	}
	return s.Cron(fmt.Sprintf("%s %s * * *", tt[0], tt[1]))
}

// Weekly 每周运行job，在周一00:00:00运行
func (s spec) Weekly() (task.TaskID, error) {
	return s.Cron("@weekly")
}

// Monthly 每月运行job，在1号00:00:00运行
func (s spec) Monthly() (task.TaskID, error) {
	return s.Cron("@monthly")
}
