package log

import (
	"context"
	stdLog "github.com/go-kratos/kratos/v2/log"
	//"github.com/go-kratos/kratos/v2/middleware/tracing"
	nativeZap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LevelDebug = stdLog.LevelDebug
	LevelInfo  = stdLog.LevelInfo
	LevelWarn  = stdLog.LevelWarn
	LevelError = stdLog.LevelError
	LevelFatal = stdLog.LevelFatal
)

type (
	Valuer = stdLog.Valuer
	Level  = stdLog.Level
)

type Logger interface {
	// AddFilter 添加过滤器。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
	AddFilter(option stdLog.FilterOption) Logger
	// AddValuer 添加Key-Valuer。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
	AddValuer(keyVals ...interface{}) Logger
	// AddStack 增加stack。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
	AddStack(skip int) Logger
	ZapCore() zapcore.Core
	Build() stdLog.Logger
	SetLevel(level string) Logger
	// Clone 克隆一个新的logger，后续使用需要先Build
	Clone() Logger

	stdLog.Logger
}

// zapLogger 日志扩展实例
type zapLogger struct {
	nativeZapCore zapcore.Core

	filters []stdLog.FilterOption
	valuers []interface{}

	stack        int
	kratosLogger stdLog.Logger
	baseContext  context.Context
}

var DefaultLogger Logger = (*zapLogger)(nil)

// Default 实例化默认日志
func Default(baseCtx context.Context, opts ...ZapCoreOption) Logger {
	kvs := []interface{}{"ts", DefaultTimestamp /*, "call", SimpleCaller(7)*/}
	DefaultLogger = New(baseCtx, opts...).AddValuer(kvs...)
	return DefaultLogger
}

// New 实例化日志，默认带有trace.id和span.id
func New(baseCtx context.Context, opts ...ZapCoreOption) Logger {
	/**
	堆栈：
	zap.(*Logger).Log (zap.go:31) github.com/go-kratos/kratos/contrib/log/zap/v2
	log.(*logger).Log (log.go:30) github.com/go-kratos/kratos/v2/log
	log.(*Filter).Log (filter.go:95) github.com/go-kratos/kratos/v2/log   <--- 这行在下面的filters中有加上
	log.(*Helper).Info (helper.go:120) kratos-packages/pkg/log
	*/
	return &zapLogger{
		nativeZapCore: buildZapCore(opts...),
		stack:         3,
		valuers:       []any{
			//"trace.id", tracing.TraceID(),
			//"span.id", tracing.SpanID(),
		},
		baseContext: baseCtx,
	}
}

// ZapCore 获取原生zap core
func (l *zapLogger) ZapCore() zapcore.Core {
	return l.nativeZapCore
}

// SetLevel 设置日志级别
func (l *zapLogger) SetLevel(level string) Logger {
	l.AddFilter(stdLog.FilterLevel(stdLog.ParseLevel(level)))
	return l
}

// AddValuer 添加Key-Valuer。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
//
//	keyvals... 为偶数个, key为字符串, val为Valuer
func (l *zapLogger) AddValuer(keyVals ...interface{}) Logger {
	l.valuers = append(l.valuers, keyVals...)
	return l
}

// AddFilter 添加过滤器。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
func (l *zapLogger) AddFilter(option stdLog.FilterOption) Logger {
	l.filters = append(l.filters, option)
	return l
}

// AddStack 增加stack。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
func (l *zapLogger) AddStack(i int) Logger {
	l.stack += i
	return l
}

func (l *zapLogger) Build() stdLog.Logger {
	zLogger := nativeZap.New(l.nativeZapCore, nativeZap.WithCaller(true), nativeZap.AddCallerSkip(len(l.filters)+l.stack))
	kratosZapLogger := zap.NewLogger(zLogger)
	// 这里的执行顺序不能变，先filters，再valuers，否则即使WithContext，valuers执行时ctx参数还是background
	kratosLogger := stdLog.With(
		stdLog.NewFilter(kratosZapLogger, l.filters...),
		l.valuers...)
	// 附加基础的context
	kratosLogger = stdLog.WithContext(l.baseContext, kratosLogger)

	return kratosLogger
}

func (l *zapLogger) Log(level Level, keyVals ...interface{}) error {
	// 如果是从这个方法进来的，走的是官方的日志接口。
	// 需要创建新的kratosLogger，并且stack需要加1
	if l.kratosLogger == nil {
		l.kratosLogger = l.Clone().AddStack(1).Build()
	}
	return l.kratosLogger.Log(level, keyVals...)
}

func (l *zapLogger) Clone() Logger {
	return &zapLogger{
		nativeZapCore: l.nativeZapCore,
		filters:       append([]stdLog.FilterOption{}, l.filters...),
		valuers:       append([]any{}, l.valuers...),
		stack:         l.stack,
		kratosLogger:  nil,
		baseContext:   l.baseContext,
	}
}
