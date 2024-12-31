package log

import (
	kratosLog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap/zapcore"
)

const (
	LevelDebug = kratosLog.LevelDebug
	LevelInfo  = kratosLog.LevelInfo
	LevelWarn  = kratosLog.LevelWarn
	LevelError = kratosLog.LevelError
	LevelFatal = kratosLog.LevelFatal
)

type (
	Valuer = kratosLog.Valuer
	Level  = kratosLog.Level
)

type Logger interface {
	// AddFilter 添加过滤器。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
	AddFilter(option kratosLog.FilterOption) Logger
	// AddValuer 添加Key-Valuer。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
	AddValuer(keyVals ...any) Logger
	// AddStack 增加stack。修改的是当前的logger，在log.NewModuleHelper中设置时，请Clone后使用。
	AddStack(skip int) Logger
	ZapCore() zapcore.Core
	Build() kratosLog.Logger
	// Clone 克隆一个新的logger，后续使用需要先Build
	Clone() Logger

	kratosLog.Logger
}
