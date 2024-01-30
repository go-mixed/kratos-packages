package log

import (
	"io"

	stdLog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap/zapcore"
)

const (
	// DefaultRotationMaxAge 默认最大日志切割生命周期
	DefaultRotationMaxAge = 30

	// DefaultRotationMaxSize 默认最大切割日志大小
	DefaultRotationMaxSize = 100 // megabytes

	// DefaultRotationMaxBackups 默认切割日志最大备份数量
	DefaultRotationMaxBackups = 3
)

type RotateOption struct {
	dir              string
	maxSize          int
	maxAge           int
	maxBackups       int
	localTime        bool
	compress         bool
	multiLevelOutput bool
}

type ZapCoreOption func(l *ZLog)

// WithRotate 是否开启日志切割
func WithRotate(rotation bool) ZapCoreOption {
	return func(l *ZLog) {
		l.rotation = rotation
	}
}

// WithMultiLevelOutput 是否开启多等级日志输出指定文件
func WithMultiLevelOutput(multi bool) ZapCoreOption {
	return func(l *ZLog) {
		l.rotateOpts.multiLevelOutput = multi
	}
}

// WithRotateLocalTime 日志切割是否使用本地时间
func WithRotateLocalTime() ZapCoreOption {
	return func(l *ZLog) {
		l.rotateOpts.localTime = true
	}
}

// WithRotateCompress 切割日志是否压缩
func WithRotateCompress() ZapCoreOption {
	return func(l *ZLog) {
		l.rotateOpts.compress = true
	}
}

// WithRotateMaxSize 最大切割日志大小
func WithRotateMaxSize(maxSize int) ZapCoreOption {
	return func(l *ZLog) {
		l.rotateOpts.maxSize = maxSize
	}
}

// WithRotateMaxAge 切割日志最大生命周期
func WithRotateMaxAge(maxAge int) ZapCoreOption {
	return func(l *ZLog) {
		l.rotateOpts.maxAge = maxAge
	}
}

// WithRotateMaxBackups 切割日志最大备份数量
func WithRotateMaxBackups(backups int) ZapCoreOption {
	return func(l *ZLog) {
		l.rotateOpts.maxBackups = backups
	}
}

// WithRotateDir 定义切割日志存放目录
func WithRotateDir(dirPath string) ZapCoreOption {
	return func(l *ZLog) {
		l.rotateOpts.dir = dirPath
	}
}

// WithColor 是否开启彩色控制台输出
func WithColor(color bool) ZapCoreOption {
	return func(l *ZLog) {
		l.color = color
	}
}

// WithWriter 添加自定义文件写入
func WithWriter(w io.Writer) ZapCoreOption {
	return func(l *ZLog) {
		l.writers = append(l.writers, w)
	}
}

// WithProduction 是否开启生产, 开启后日志使用json输出
func WithProduction(production bool) ZapCoreOption {
	return func(l *ZLog) {
		l.production = production
	}
}

// WithLevel 设置日志级别
func WithLevel(level string) ZapCoreOption {
	return func(l *ZLog) {
		stdLevel := stdLog.ParseLevel(level)
		switch stdLevel {
		case stdLog.LevelDebug:
			l.level = zapcore.DebugLevel
		case stdLog.LevelInfo:
			l.level = zapcore.InfoLevel
		case stdLog.LevelWarn:
			l.level = zapcore.WarnLevel
		case stdLog.LevelError:
			l.level = zapcore.ErrorLevel
		case stdLog.LevelFatal:
			l.level = zapcore.FatalLevel
		default:
			l.level = zapcore.InfoLevel
		}
	}
}
