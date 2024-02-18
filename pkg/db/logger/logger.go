package logger

import (
	"context"
	"errors"
	"fmt"
	kratosLog "gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"time"

	gormLogger "gorm.io/gorm/logger"
)

// New initialize logger
func New(logHelper *kratosLog.Helper, config Config) Interface {
	var (
		traceStr     = "[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		traceStr = gormLogger.Yellow + "[%.3fms] " + gormLogger.BlueBold + "[rows:%v]" + gormLogger.Reset + " %s"
		traceWarnStr = gormLogger.Yellow + "%s\n" + gormLogger.Reset + gormLogger.RedBold + "[%.3fms] " + gormLogger.Yellow + "[rows:%v]" + gormLogger.Magenta + " %s" + gormLogger.Reset
		traceErrStr = gormLogger.MagentaBold + "%s\n" + gormLogger.Reset + gormLogger.Yellow + "[%.3fms] " + gormLogger.BlueBold + "[rows:%v]" + gormLogger.Reset + " %s"
	}

	return &logger{
		Config:       config,
		kratosLogger: logHelper,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type logger struct {
	Config
	kratosLogger                        *kratosLog.Helper
	traceStr, traceErrStr, traceWarnStr string
}

var _ Interface = (*logger)(nil)

// LogMode log mode
func (l *logger) LogMode(level gormLogger.LogLevel) Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *logger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Info {
		l.kratosLogger.WithContext(ctx).Infof(msg, data...)
	}
}

// Warn print warn messages
func (l *logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Warn {
		l.kratosLogger.WithContext(ctx).Warnf(msg, data...)
	}
}

// Error print error messages
func (l *logger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormLogger.Error {
		l.kratosLogger.WithContext(ctx).Errorf(msg, data...)
	}
}

// Trace print sql message
func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.kratosLogger.WithContext(ctx).Errorf(l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.kratosLogger.WithContext(ctx).Errorf(l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.kratosLogger.WithContext(ctx).Warnf(l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.kratosLogger.WithContext(ctx).Warnf(l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.kratosLogger.WithContext(ctx).Debugf(l.traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.kratosLogger.WithContext(ctx).Debugf(l.traceStr, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// Trace print sql message
func (l *logger) ParamsFilter(ctx context.Context, sql string, params ...any) (string, []any) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
