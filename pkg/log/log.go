package log

import (
	"context"
	"github.com/go-kratos/kratos/contrib/log/zap/v2"
	stdLog "github.com/go-kratos/kratos/v2/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/config"
	"time"

	//"github.com/go-kratos/kratos/v2/middleware/tracing"
	nativeZap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger 日志扩展实例
type zapLogger struct {
	nativeZapCore zapcore.Core

	filters []stdLog.FilterOption
	valuers []any

	stack        int
	kratosLogger stdLog.Logger
	baseContext  context.Context
}

var DefaultLogger Logger = (*zapLogger)(nil)

// Default 实例化默认日志
func Default(baseCtx context.Context, opts ...zapSimpleOption) Logger {
	kvs := []any{"ts", DefaultTimestamp /*, "call", SimpleCaller(7)*/}
	DefaultLogger = NewSimple(baseCtx, opts...).AddValuer(kvs...)
	return DefaultLogger
}

// NewFromConfig 从完整的配置创建日志实例
func NewFromConfig(baseCtx context.Context, configure config.Configure) Logger {
	var logConf logConfig
	// 映射配置
	if err := configure.Value("logger").Scan(&logConf); err != nil {
		panic(err)
	}

	for i, writer := range logConf.Writers {
		// 设置levelValue
		if writer.Level == "" {
			writer.levelValue = LevelDebug
		} else {
			writer.levelValue = stdLog.ParseLevel(writer.Level)
		}
		// 设置时间格式
		if writer.TimeFormat == "" {
			writer.TimeFormat = time.RFC3339
		}
		// write变量是副本，需要重新赋值到logConf.Writers
		logConf.Writers[i] = writer
	}

	/**
	堆栈：有filters时，stack=4，没有filters时，stack=3
	zap.(*Logger).Log (zap.go:38) github.com/go-kratos/kratos/contrib/log/zap/v2
	log.(*Filter).Log (filter.go:93) github.com/go-kratos/kratos/v2/log  <-- 如果没有设置filters，就没有这一层
	log.(*logger).Log (log.go:30) github.com/go-kratos/kratos/v2/log
	log.(*Helper).Info (helper.go:158) gopkg.in/go-mixed/kratos-packages.v2/pkg/log
	*/
	return &zapLogger{
		nativeZapCore: buildZapCore(logConf),
		stack:         3,
		valuers:       []any{
			//"trace.id", tracing.TraceID(),
			//"span.id", tracing.SpanID(),
		},
		filters:     []stdLog.FilterOption{stdLog.FilterLevel(stdLog.LevelDebug)},
		baseContext: baseCtx,
	}
}

// NewSimple 实例化简单的日志，默认带有trace.id和span.id
func NewSimple(baseCtx context.Context, opts ...zapSimpleOption) Logger {
	conf := &simpleLogConf{
		level: zapcore.DebugLevel,
	}

	for _, opt := range opts {
		opt(conf)
	}
	/**
	堆栈：有filters时，stack=4，没有filters时，stack=3
	zap.(*Logger).Log (zap.go:38) github.com/go-kratos/kratos/contrib/log/zap/v2
	log.(*Filter).Log (filter.go:93) github.com/go-kratos/kratos/v2/log  <-- 如果没有设置filters，就没有这一层
	log.(*logger).Log (log.go:30) github.com/go-kratos/kratos/v2/log
	log.(*Helper).Info (helper.go:158) gopkg.in/go-mixed/kratos-packages.v2/pkg/log
	*/
	return &zapLogger{
		nativeZapCore: buildSimpleZapCore(*conf),
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

//// SetLevel 设置日志级别
//func (l *zapLogger) SetLevel(level string) Logger {
//	l.AddFilter(stdLog.FilterLevel(stdLog.ParseLevel(level)))
//	return l
//}

// AddValuer 添加Key-Valuer。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
//
//	keyvals... 为偶数个, key为字符串, val为Valuer
func (l *zapLogger) AddValuer(keyVals ...any) Logger {
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
	var filters []stdLog.FilterOption
	stack := l.stack
	if len(l.filters) > 0 {
		// 多1层调用链：log.(*Filter).Log (filter.go:93) github.com/go-kratos/kratos/v2/log
		stack += 1
		// 如果有filter，则必须初始化一个Level的filter，不然在 Filter.Log 阶段就会被拦截掉，
		// 这里写死为debug，放行所有日志，后面的zap的level拦截器会拦截掉。
		filters = append([]stdLog.FilterOption{stdLog.FilterLevel(stdLog.LevelDebug)}, l.filters...)
	}
	zLogger := nativeZap.New(l.nativeZapCore, nativeZap.WithCaller(true), nativeZap.AddCallerSkip(stack))
	var kratosZapLogger stdLog.Logger = zap.NewLogger(zLogger)
	// 这里的执行顺序不能变，先filters，再valuers，否则在WithContext之后，执行时ctx参数还是background
	if len(filters) > 0 {
		kratosZapLogger = stdLog.NewFilter(kratosZapLogger, filters...)
	}
	// 附加基础的context
	return stdLog.WithContext(l.baseContext, stdLog.With(kratosZapLogger, l.valuers...))
}

func (l *zapLogger) Log(level Level, keyVals ...any) error {
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
