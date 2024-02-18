package schedule

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/job"
	"math"
	"strings"
	"time"
)

type Spec interface {
	Cron(expression any) (cron.EntryID, error)
	Every(duration time.Duration) (cron.EntryID, error)
	EverySeconds(seconds ...int) (cron.EntryID, error)
	EveryMinute() (cron.EntryID, error)
	AfterEveryMinute() (cron.EntryID, error)
	EveryMinutes(minutes int) (cron.EntryID, error)
	AfterEveryMinutes(minutes int) (cron.EntryID, error)
	EveryFiveMinutes() (cron.EntryID, error)
	AfterEveryFiveMinutes() (cron.EntryID, error)
	EveryTenMinutes() (cron.EntryID, error)
	AfterEveryTenMinutes() (cron.EntryID, error)
	EveryFifteenMinutes() (cron.EntryID, error)
	AfterEveryFifteenMinutes() (cron.EntryID, error)
	EveryThirtyMinutes() (cron.EntryID, error)
	AfterEveryThirtyMinutes() (cron.EntryID, error)
	Hourly() (cron.EntryID, error)
	HourlyAt(offset int) (cron.EntryID, error)
	Daily() (cron.EntryID, error)
	DailyAt(t string) (cron.EntryID, error)
	Weekly() (cron.EntryID, error)
	Monthly() (cron.EntryID, error)
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

type cronCaller func(spec any, job job.Job) (cron.EntryID, error)

type spec struct {
	job    job.Job
	runner cronCaller
}

// NewSpec 实例化cron表达式
func NewSpec(runner cronCaller, job job.Job) *spec {
	return &spec{
		job:    job,
		runner: runner,
	}
}

// Cron 自定义cron表达式运行job
// 支持表达式：https://pkg.go.dev/github.com/robfig/cron/v3#hdr-Special_Characters
func (s spec) Cron(spec any) (cron.EntryID, error) {
	return s.runner(spec, s.job)
}

// Every 传入time.Duration运行job，精度为秒
func (s spec) Every(duration time.Duration) (cron.EntryID, error) {
	return s.runner(cron.Every(duration), s.job)
}

// EverySeconds 每秒运行job
func (s spec) EverySeconds(seconds ...int) (cron.EntryID, error) {
	defaultSeconds := 1
	if len(seconds) > 0 {
		defaultSeconds = seconds[0]
	}
	return s.Cron(fmt.Sprintf("@every %ds", defaultSeconds))
}

// EveryMinute 每分钟运行job
func (s spec) EveryMinute() (cron.EntryID, error) {
	return s.EveryMinutes(1)
}

// AfterEveryMinute 每分钟后运行job
func (s spec) AfterEveryMinute() (cron.EntryID, error) {
	return s.Cron("@every 1m")
}

// EveryMinutes 每多少分钟运行job
func (s spec) EveryMinutes(minutes int) (cron.EntryID, error) {
	return s.Cron(fmt.Sprintf("0 0/%d * * * *", minutes))
}

// AfterEveryMinutes 每多少分钟后运行job
func (s spec) AfterEveryMinutes(minutes int) (cron.EntryID, error) {
	return s.Cron(fmt.Sprintf("@every %dm", minutes))
}

// EveryFiveMinutes 每五分钟运行job
func (s spec) EveryFiveMinutes() (cron.EntryID, error) {
	return s.EveryMinutes(5)
}

// AfterEveryFiveMinutes 每五分钟后运行job
func (s spec) AfterEveryFiveMinutes() (cron.EntryID, error) {
	return s.AfterEveryMinutes(5)
}

// EveryTenMinutes 每十分钟运行job
func (s spec) EveryTenMinutes() (cron.EntryID, error) {
	return s.EveryMinutes(10)
}

// AfterEveryTenMinutes 每十分钟后运行job
func (s spec) AfterEveryTenMinutes() (cron.EntryID, error) {
	return s.AfterEveryMinutes(10)
}

// EveryFifteenMinutes 每十五分钟运行job
func (s spec) EveryFifteenMinutes() (cron.EntryID, error) {
	return s.EveryMinutes(15)
}

// AfterEveryFifteenMinutes 每十五分钟后运行job
func (s spec) AfterEveryFifteenMinutes() (cron.EntryID, error) {
	return s.AfterEveryMinutes(15)
}

// EveryThirtyMinutes 每三十分钟运行job
func (s spec) EveryThirtyMinutes() (cron.EntryID, error) {
	return s.EveryMinutes(30)
}

// AfterEveryThirtyMinutes 每三十分钟后运行job
func (s spec) AfterEveryThirtyMinutes() (cron.EntryID, error) {
	return s.AfterEveryMinutes(30)
}

// Hourly 每小时运行job
func (s spec) Hourly() (cron.EntryID, error) {
	return s.Cron("@hourly")
}

// HourlyAt 每小时的某分钟运行job
func (s spec) HourlyAt(offset int) (cron.EntryID, error) {
	return s.Cron(fmt.Sprintf("@every 1h%dm", offset))
}

// Daily 每天运行job
func (s spec) Daily() (cron.EntryID, error) {
	return s.Cron("@daily")
}

// DailyAt 每天某时某分运行job
// DailyAt("12:21") 每天12点21分钟运行
func (s spec) DailyAt(t string) (cron.EntryID, error) {
	tt := strings.Split(t, ":")
	if len(tt) < 2 {
		tt = append(tt, "00")
	}
	return s.Cron(fmt.Sprintf("%s %s * * *", tt[0], tt[1]))
}

// Weekly 每周运行job
func (s spec) Weekly() (cron.EntryID, error) {
	return s.Cron("@weekly")
}

// Monthly 每月运行job
func (s spec) Monthly() (cron.EntryID, error) {
	return s.Cron("@monthly")
}
