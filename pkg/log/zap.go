package log

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"time"
)

func buildZapEncoder(encoderType encoderType, timeFormat string, color bool) zapcore.Encoder {
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}
	zConf := zapcore.EncoderConfig{
		TimeKey:     "time",
		LevelKey:    "level",
		NameKey:     "log",
		CallerKey:   "caller",
		FunctionKey: zapcore.OmitKey,
		MessageKey:  "msg", // 因为是 Helper 唤起日志，并且helper中包含的msg字段，所以这里设置为空
		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format(timeFormat))
		},
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.FullCallerEncoder, // 显示caller的完整文件路径
		ConsoleSeparator: " ",
	}

	var encoder zapcore.Encoder
	switch encoderType {
	case encoderTypeJSON:
		encoder = zapcore.NewJSONEncoder(zConf)
	case encoderTypeConsole:
		fallthrough
	default: // 默认为console
		if color {
			zConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(zConf)
	}

	return encoder
}

func buildSyncer(outputType outputType, fileConfig fileConfig) zapcore.WriteSyncer {
	switch outputType {
	case OutputTypeStderr:
		return zapcore.AddSync(os.Stderr)
	case OutputTypeFile:
		return buildFieSyncer(fileConfig)
	case OutputTypeFluent, OutputTypeAliyun, OutputTypeTencent:
		panic("不支持的日志输出类型: " + outputType)
	case OutputTypeStdout:
		fallthrough
	default: // 默认为stdout
		return zapcore.AddSync(os.Stdout)
	}
}

func buildFieSyncer(config fileConfig) zapcore.WriteSyncer {
	if config.Path == "" {
		panic("must provide a rotation dir path")
	}
	// 创建目录
	_ = os.MkdirAll(filepath.Dir(config.Path), os.ModePerm)

	// 不启用日志轮转，则直接输出到文件
	if !config.Rotate {
		fs, err := os.OpenFile(config.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			println("write log file failed: " + err.Error())
			return zapcore.AddSync(io.Discard) // 输出到空
		}
		return zapcore.AddSync(fs)
	}

	if config.MaxAge == 0 {
		config.MaxAge = DefaultRotationMaxAge
	}
	if config.MaxSize == 0 {
		config.MaxSize = DefaultRotationMaxSize
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = DefaultRotationMaxBackups
	}
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   config.Path,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		LocalTime:  config.LocalTime,
		Compress:   config.Compress,
	})
}

// buildFilterFunc 构建过滤器
func buildFilterFunc(filters []FilterOption) log.FilterOption {
	return log.FilterFunc(func(level log.Level, kvs ...interface{}) bool {
		for _, filter := range filters {
			if !filter.Enable {
				continue
			}
			for i := 0; i < len(kvs); i++ {
				keyword := kvs[i]
				if filter.MatchValue {
					keyword = kvs[i+1]
				}
				if k, ok := keyword.(string); ok {
					if filter.MatchRule.Match()(k, filter.Target) {
						if filter.Ignore != nil && *filter.Ignore {
							return true
						}
						if filter.Replace != nil {
							if i+1 < len(kvs) {
								kvs[i+1] = filter.Replace
							} else {
								kvs[i] = filter.Target + ":" + *filter.Replace
							}
						}
					}
				}
			}
		}
		return false
	})
}

// buildSimpleZapCore 实例化简单的Zap Core，绝对会输出到stdout、文件（logger.log）；如果conf.multiLevelOutput为true，则输出到stdout、logger.log、info.log、warn.log
//
//	simpleLogConf.dir 为日志目录
func buildSimpleZapCore(conf simpleLogConf) zapcore.Core {
	atomic := zap.NewAtomicLevelAt(conf.level)
	encoder := buildZapEncoder(lo.If(conf.production, encoderTypeJSON).Else(encoderTypeConsole), "2006-01-02 15:04:05", conf.color)

	syncers := []zapcore.WriteSyncer{
		zapcore.AddSync(os.Stdout),
	}

	// 如果设置了日志目录，则添加文件的syncer
	if conf.dir != "" {
		syncers = append(syncers, buildFieSyncer(fileConfig{filepath.Join(conf.dir, "logger.log"), conf.rotateConfig.Rotate, conf.rotateConfig.MaxSize, conf.rotateConfig.MaxAge, conf.rotateConfig.MaxBackups, conf.rotateConfig.LocalTime, conf.rotateConfig.Compress}))
	}

	coreTee := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(syncers...), atomic),
	}

	if conf.dir != "" && conf.multiLevelOutput {
		infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.WarnLevel && lvl >= atomic.Level()
		})

		warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.WarnLevel && lvl >= atomic.Level()
		})

		coreTee = append(coreTee, []zapcore.Core{
			zapcore.NewCore(encoder, zapcore.AddSync(
				buildFieSyncer(fileConfig{filepath.Join(conf.dir, "info.log"), conf.rotateConfig.Rotate, conf.rotateConfig.MaxSize, conf.rotateConfig.MaxAge, conf.rotateConfig.MaxBackups, conf.rotateConfig.LocalTime, conf.rotateConfig.Compress}),
			), infoLevel),
			zapcore.NewCore(encoder, zapcore.AddSync(
				buildFieSyncer(fileConfig{filepath.Join(conf.dir, "warn.log"), conf.rotateConfig.Rotate, conf.rotateConfig.MaxSize, conf.rotateConfig.MaxAge, conf.rotateConfig.MaxBackups, conf.rotateConfig.LocalTime, conf.rotateConfig.Compress}),
			), warnLevel),
		}...)
	}

	return zapcore.NewTee(coreTee...)
}

// buildZapCore 实例化Zap Core，支持多个writer
//
//	logConfig.Writers.File.Path 设置的是日志文件的路径，这是和buildSimpleZapCore不同的地方。如果要根据Level输出多个文件，需要设置多个writer
func buildZapCore(conf logConfig) zapcore.Core {
	var coreTee []zapcore.Core
	for _, writer := range conf.Writers {
		if !writer.Enable {
			continue
		}
		encoder := buildZapEncoder(writer.Encoder, writer.TimeFormat, writer.Color)
		atomic := zap.NewAtomicLevelAt(zapcore.Level(writer.levelValue))
		syncer := buildSyncer(writer.Output, writer.File)

		core := zapcore.NewCore(encoder, syncer, atomic)
		coreTee = append(coreTee, core)
	}

	return zapcore.NewTee(coreTee...)
}
