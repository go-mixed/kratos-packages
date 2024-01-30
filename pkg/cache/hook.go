package cache

import (
	"context"
	stdRedis "github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"strings"
)

type cacheHook struct {
	options        redis.Options
	logger         *log.Helper
	pipelineLogger *log.Helper
}

func newCacheHook(options redis.Options, logger log.Logger) *cacheHook {
	return &cacheHook{
		options:        options,
		logger:         log.NewModuleHelper(logger.Clone().AddStack(5), "cache", "duration", LogValuer()),
		pipelineLogger: log.NewModuleHelper(logger.Clone().AddStack(8), "cache", "duration", LogValuer()),
	}
}

var _ stdRedis.Hook = (*cacheHook)(nil)

func (k cacheHook) BeforeProcess(ctx context.Context, cmd stdRedis.Cmder) (context.Context, error) {
	ctx = NewContext(ctx)
	return ctx, nil
}

func (k cacheHook) AfterProcess(ctx context.Context, cmd stdRedis.Cmder) error {
	// 打印所有的redis操作
	k.logger.WithContext(ctx).Debugf("[Redis] %s", cmd.String())
	return nil
}

func (k cacheHook) BeforeProcessPipeline(ctx context.Context, cmds []stdRedis.Cmder) (context.Context, error) {
	ctx = NewContext(ctx)
	return ctx, nil
}

func (k cacheHook) AfterProcessPipeline(ctx context.Context, cmds []stdRedis.Cmder) error {
	str := strings.Join(lo.Map(cmds, func(cmd stdRedis.Cmder, _ int) string {
		return cmd.String()
	}), "\n")
	k.pipelineLogger.WithContext(ctx).Debugf("[RedisPipeline] %s", str)
	return nil
}

func (k cacheHook) DialHook(next stdRedis.DialHook) stdRedis.DialHook {
	return nil
}

func (k cacheHook) ProcessHook(next stdRedis.ProcessHook) stdRedis.ProcessHook {
	return nil
}

func (k cacheHook) ProcessPipelineHook(next stdRedis.ProcessPipelineHook) stdRedis.ProcessPipelineHook {
	return nil
}
