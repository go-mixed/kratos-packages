package log

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2"
	stdLog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/pkg/errors"
	gormLogger "gorm.io/gorm/logger"
	"os"
	"strconv"
)

// DefaultMessageKey default message key.
var DefaultMessageKey = "msg"

// Option is Helper option.
type Option func(*Helper)

// Helper is a logger helper.
type Helper struct {
	logger stdLog.Logger
	msgKey string
}

// Need to implement gormLogger.Writer interface
var _ gormLogger.Writer = (*Helper)(nil)

// WithMessageKey with message key.
func WithMessageKey(k string) Option {
	return func(opts *Helper) {
		opts.msgKey = k
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// LogValuer 提供日志valuer
func appIDLogValuer() stdLog.Valuer {
	return func(ctx context.Context) interface{} {
		if app, ok := kratos.FromContext(ctx); ok {
			return app.ID()
		}
		return ""
	}
}

// NewModuleHelper 实例化模块日志助手
func NewModuleHelper(l Logger, moduleName string, kv ...any) *Helper {
	kv = append(kv,
		"module", moduleName,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
		"app.id", appIDLogValuer(),
	)

	l = l.Clone().AddValuer(kv...)

	return newHelper(l)
}

// newHelper new a logger helper.
func newHelper(logger Logger, opts ...Option) *Helper {
	options := &Helper{
		msgKey: DefaultMessageKey, // default message key
		logger: logger.Build(),
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

func stackToLines(stack stackTracer) []string {
	var lines []string
	for _, f := range stack.StackTrace() {
		lines = append(lines, fmt.Sprintf("%+s:%d", f, f))
	}

	return lines
}

func (h *Helper) getStack(keyVals ...any) []any {
	var kvs []any
	// for each keyVals, if it is an error of stackTracer, then add stack trace to the log
	for i := 0; i < len(keyVals); i++ {
		if st, ok := keyVals[i].(stackTracer); ok {
			kvs = append(kvs, "error-"+strconv.Itoa(i), stackToLines(st))
		}
	}

	return kvs
}

func (h *Helper) rawWithStack(keyVals ...any) []any {
	return append(keyVals, h.getStack(keyVals...)...)
}

func (h *Helper) sprintWithStack(keyVals ...any) []any {
	return append([]any{h.msgKey, fmt.Sprint(keyVals...)}, h.getStack(keyVals...)...)
}

func (h *Helper) sprintfWithStack(format string, keyVals ...any) []any {
	return append([]any{h.msgKey, fmt.Sprintf(format, keyVals...)}, h.getStack(keyVals...)...)
}

// WithContext returns a shallow copy of h with its context changed
// to ctx. The provided ctx must be non-nil.
func (h *Helper) WithContext(ctx context.Context) *Helper {
	return &Helper{
		msgKey: h.msgKey,
		logger: stdLog.WithContext(ctx, h.logger),
	}
}

// Log Print log by level and keyvals.
func (h *Helper) Log(level Level, keyvals ...any) error {
	return h.logger.Log(level, keyvals...)
}

// Debug logs a message at debug level.
func (h *Helper) Debug(a ...any) {
	_ = h.logger.Log(LevelDebug, h.sprintWithStack(a...)...)
}

// Debugf logs a message at debug level.
func (h *Helper) Debugf(format string, a ...any) {
	_ = h.logger.Log(LevelDebug, h.sprintfWithStack(format, a...)...)
}

// Debugw logs a message at debug level.
func (h *Helper) Debugw(keyvals ...any) {
	_ = h.logger.Log(LevelDebug, h.rawWithStack(keyvals...)...)
}

// Info logs a message at info level.
func (h *Helper) Info(a ...any) {
	_ = h.logger.Log(LevelInfo, h.sprintWithStack(a...)...)
}

// Infof logs a message at info level.
func (h *Helper) Infof(format string, a ...any) {
	_ = h.logger.Log(LevelInfo, h.sprintfWithStack(format, a...)...)
}

// Infow logs a message at info level.
func (h *Helper) Infow(keyvals ...any) {
	_ = h.logger.Log(LevelInfo, h.rawWithStack(keyvals...)...)
}

// Warn logs a message at warn level.
func (h *Helper) Warn(a ...any) {
	_ = h.logger.Log(LevelWarn, h.sprintWithStack(a...)...)
}

// Warnf logs a message at warnf level.
func (h *Helper) Warnf(format string, a ...any) {
	_ = h.logger.Log(LevelWarn, h.sprintfWithStack(format, a...)...)
}

// Warnw logs a message at warnf level.
func (h *Helper) Warnw(keyvals ...any) {
	_ = h.logger.Log(LevelWarn, h.rawWithStack(keyvals...)...)
}

// Error logs a message at error level.
func (h *Helper) Error(a ...any) {
	_ = h.logger.Log(LevelError, h.sprintWithStack(a...)...)
}

// Errorf logs a message at error level.
func (h *Helper) Errorf(format string, a ...any) {
	_ = h.logger.Log(LevelError, h.sprintfWithStack(format, a...)...)
}

// Errorw logs a message at error level.
func (h *Helper) Errorw(keyvals ...any) {
	_ = h.logger.Log(LevelError, h.rawWithStack(keyvals...)...)
}

// Fatal logs a message at fatal level.
func (h *Helper) Fatal(a ...any) {
	_ = h.logger.Log(LevelFatal, h.sprintWithStack(a...)...)
	os.Exit(1)
}

// Fatalf logs a message at fatal level.
func (h *Helper) Fatalf(format string, a ...any) {
	_ = h.logger.Log(LevelFatal, h.sprintfWithStack(format, a...)...)
	os.Exit(1)
}

// Fatalw logs a message at fatal level.
func (h *Helper) Fatalw(keyvals ...any) {
	_ = h.logger.Log(LevelFatal, h.rawWithStack(keyvals...)...)
	os.Exit(1)
}

// Printf 提供给Gorm使用的日志接口
func (h *Helper) Printf(format string, a ...any) {
	_ = h.logger.Log(LevelInfo, h.sprintfWithStack(format, a...)...)
}
