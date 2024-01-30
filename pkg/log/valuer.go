package log

import (
	"context"
	stdLog "github.com/go-kratos/kratos/v2/log"
	"runtime"
	"strconv"
)

var (
	DefaultTimestamp = stdLog.DefaultTimestamp
)

// SimpleCaller 简单caller
func SimpleCaller(depth int) Valuer {
	return stdLog.Caller(depth)
}

// FullCaller 完成输出caller
func FullCaller(depth int) Valuer {
	return func(ctx context.Context) interface{} {
		_, file, line, _ := runtime.Caller(depth)
		return file + ":" + strconv.Itoa(line)
	}
}
