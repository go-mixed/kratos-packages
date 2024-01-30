package schedule

import (
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

type scheduleLogger struct {
	h *log.Helper
}

var _ cron.Logger = (*scheduleLogger)(nil)

func NewScheduleLogger(logger log.Logger) *scheduleLogger {
	return &scheduleLogger{
		h: log.NewModuleHelper(logger.Clone().AddStack(1), "server/schedule"),
	}
}

func (s scheduleLogger) Info(msg string, keysAndValues ...any) {
	s.h.Infof(msg, keysAndValues...)
}

func (s scheduleLogger) Error(err error, msg string, keysAndValues ...any) {
	s.h.Errorf(msg, append([]any{err}, keysAndValues...)...)
}
