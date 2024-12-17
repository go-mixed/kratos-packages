package log

import (
	stdLog "github.com/go-kratos/kratos/v2/log"
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
	AddValuer(keyVals ...any) Logger
	// AddStack 增加stack。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
	AddStack(skip int) Logger
	ZapCore() zapcore.Core
	Build() stdLog.Logger
	// SetLevel(level string) Logger
	// Clone 克隆一个新的logger，后续使用需要先Build
	Clone() Logger

	stdLog.Logger
}
