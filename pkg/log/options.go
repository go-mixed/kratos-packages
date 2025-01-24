package log

import (
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

type zapSimpleOption func(l *simpleLogConf)

// WithRotate 是否开启日志切割
func WithRotate(rotation bool) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.rotateConfig.Rotate = rotation
	}
}

// WithMultiLevelOutput 是否开启多等级日志输出指定文件
func WithMultiLevelOutput(multi bool) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.multiLevelOutput = multi
	}
}

// WithRotateLocalTime 日志切割的备份文件是否以本地时间来命名，默认为UTC时间
func WithRotateLocalTime() zapSimpleOption {
	return func(l *simpleLogConf) {
		l.rotateConfig.LocalTime = true
	}
}

// WithRotateCompress 切割日志是否压缩
func WithRotateCompress() zapSimpleOption {
	return func(l *simpleLogConf) {
		l.rotateConfig.Compress = true
	}
}

// WithRotateMaxSize 最大切割日志大小（MB）
func WithRotateMaxSize(maxSize int) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.rotateConfig.MaxSize = maxSize
	}
}

// WithRotateMaxAge 切割日志最大生命周期（日）
func WithRotateMaxAge(maxAge int) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.rotateConfig.MaxAge = maxAge
	}
}

// WithRotateMaxBackups 切割日志最大备份数量（个）
func WithRotateMaxBackups(backups int) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.rotateConfig.MaxBackups = backups
	}
}

// WithRotateDir 定义切割日志存放目录
func WithRotateDir(dirPath string) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.dir = dirPath
	}
}

// WithColor 是否开启彩色控制台输出
func WithColor(color bool) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.color = color
	}
}

// WithProduction 是否开启生产, 开启后日志使用json输出
func WithProduction(production bool) zapSimpleOption {
	return func(l *simpleLogConf) {
		l.production = production
	}
}

// WithLevel 设置日志级别
func WithLevel(level string) zapSimpleOption {
	return func(l *simpleLogConf) {
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
