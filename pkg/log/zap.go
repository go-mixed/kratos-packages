package log

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZLog struct {
	core zapcore.Core

	atomic zap.AtomicLevel

	writers    []io.Writer
	color      bool
	production bool
	rotation   bool
	level      zapcore.Level

	rotateOpts RotateOption
}

// buildZapCore 实例化Zap Core
func buildZapCore(opts ...ZapCoreOption) zapcore.Core {
	l := &ZLog{
		level: zapcore.DebugLevel,
	}

	for _, opt := range opts {
		opt(l)
	}

	l.atomic = zap.NewAtomicLevelAt(l.level)

	l.init()

	return l.core
}

func (l *ZLog) init() {
	var encoder zapcore.Encoder
	conf := zapcore.EncoderConfig{
		TimeKey:    "time",
		LevelKey:   "level",
		NameKey:    "log",
		CallerKey:  "caller",
		MessageKey: "msg", // 因为是 Helper 唤起日志，并且helper中包含的msg字段，所以这里设置为空
		LineEnding: zapcore.DefaultLineEnding,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.FullCallerEncoder, // 显示caller的完整文件路径
		ConsoleSeparator: " ",
	}

	if l.color {
		conf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoder = zapcore.NewConsoleEncoder(conf)

	if l.production {
		conf.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(conf)
	}

	syncers := []zapcore.WriteSyncer{
		zapcore.AddSync(os.Stdout),
	}

	_ = os.MkdirAll(l.rotateOpts.dir, os.ModePerm)

	if l.rotation {
		syncers = append(syncers, l.getRotationSyncWriter(l.rotateOpts.dir))
	}

	for _, writer := range l.writers {
		syncers = append(syncers, zapcore.AddSync(writer))
	}

	coreTee := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(syncers...), l.atomic),
	}

	if l.rotateOpts.multiLevelOutput && l.rotation {
		infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.WarnLevel && lvl >= l.atomic.Level()
		})

		warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.WarnLevel && lvl >= l.atomic.Level()
		})

		coreTee = append(coreTee, []zapcore.Core{
			zapcore.NewCore(encoder, zapcore.AddSync(
				l.getRotationSyncWriter(l.rotateOpts.dir, "info.log"),
			), infoLevel),
			zapcore.NewCore(encoder, zapcore.AddSync(
				l.getRotationSyncWriter(l.rotateOpts.dir, "warn.log"),
			), warnLevel),
		}...)
	}

	l.core = zapcore.NewTee(coreTee...)
}

func (l *ZLog) getRotationSyncWriter(dir string, named ...string) zapcore.WriteSyncer {
	name := "logger.log"
	if dir == "" {
		panic("must provide a rotation dir path")
	}
	if len(named) > 0 {
		name = named[0]
	}
	if l.rotateOpts.maxAge == 0 {
		l.rotateOpts.maxAge = DefaultRotationMaxAge
	}
	if l.rotateOpts.maxSize == 0 {
		l.rotateOpts.maxSize = DefaultRotationMaxSize
	}
	if l.rotateOpts.maxBackups == 0 {
		l.rotateOpts.maxBackups = DefaultRotationMaxBackups
	}
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(dir, name),
		MaxSize:    l.rotateOpts.maxSize,
		MaxAge:     l.rotateOpts.maxAge,
		MaxBackups: l.rotateOpts.maxBackups,
		LocalTime:  l.rotateOpts.localTime,
		Compress:   l.rotateOpts.compress,
	})
}
