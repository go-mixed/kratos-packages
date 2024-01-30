package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type pipelinerCtx struct {
}

// newContext 将pipeliner存入context
func newContext(ctx context.Context, pipeliner redis.Pipeliner) context.Context {
	return context.WithValue(ctx, pipelinerCtx{}, pipeliner)
}

// fromContext 从context中获取pipeliner
func fromContext(ctx context.Context) (redis.Pipeliner, bool) {
	if ctx == nil {
		return nil, false
	}
	pipeliner, ok := ctx.Value(pipelinerCtx{}).(redis.Pipeliner)
	return pipeliner, ok
}
