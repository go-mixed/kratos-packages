package cache

import (
	"context"
	stdLog "github.com/go-kratos/kratos/v2/log"
	"time"
)

type cacheContext struct{}

// LogValuer 提供日志valuer
func LogValuer() stdLog.Valuer {
	return func(ctx context.Context) any {
		tm := FromContext(ctx)
		if tm.IsZero() {
			return 0
		}
		return time.Since(tm)
	}
}

func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, cacheContext{}, time.Now())
}

func FromContext(ctx context.Context) time.Time {
	value := ctx.Value(cacheContext{})
	if value == nil {
		return time.Time{}
	}
	return value.(time.Time)
}
