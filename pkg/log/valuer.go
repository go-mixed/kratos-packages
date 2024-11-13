package log

import (
	"context"
	stdLog "github.com/go-kratos/kratos/v2/log"
	"runtime"
	"strconv"
	"strings"
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
	return func(ctx context.Context) any {
		_, file, line, _ := runtime.Caller(depth)
		return file + ":" + strconv.Itoa(line)
	}
}

func Caller(depths ...int) Valuer {
	return func(context.Context) any {
		var results []string
		for _, depth := range depths {
			_, file, line, _ := runtime.Caller(depth)
			idx := strings.LastIndexByte(file, '/')
			if idx == -1 {
				return file[idx+1:] + ":" + strconv.Itoa(line)
			}
			idx = strings.LastIndexByte(file[:idx], '/')
			results = append(results, file[idx+1:]+":"+strconv.Itoa(line))
		}
		return strings.Join(results, " ")
	}
}
