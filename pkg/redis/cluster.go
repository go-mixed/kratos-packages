package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func (c *Redis) ClusterSlots(ctx context.Context) *redis.ClusterSlotsCmd {
	return c.GetRedisCmd(ctx).ClusterSlots(ctx)
}

func (c *Redis) ClusterNodes(ctx context.Context) *redis.StringCmd {
	return c.GetRedisCmd(ctx).ClusterNodes(ctx)
}

func (c *Redis) ClusterMeet(ctx context.Context, host, port string) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterMeet(ctx, host, port)
}

func (c *Redis) ClusterForget(ctx context.Context, nodeID string) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterForget(ctx, nodeID)
}

func (c *Redis) ClusterReplicate(ctx context.Context, nodeID string) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterReplicate(ctx, nodeID)
}

func (c *Redis) ClusterResetSoft(ctx context.Context) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterResetSoft(ctx)
}

func (c *Redis) ClusterResetHard(ctx context.Context) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterResetHard(ctx)
}

func (c *Redis) ClusterInfo(ctx context.Context) *redis.StringCmd {
	return c.GetRedisCmd(ctx).ClusterInfo(ctx)
}

func (c *Redis) ClusterKeySlot(ctx context.Context, key string) *redis.IntCmd {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ClusterKeySlot(ctx, key)
}

func (c *Redis) ClusterGetKeysInSlot(ctx context.Context, slot int, count int) *redis.StringSliceCmd {
	return c.GetRedisCmd(ctx).ClusterGetKeysInSlot(ctx, slot, count)
}

func (c *Redis) ClusterCountFailureReports(ctx context.Context, nodeID string) *redis.IntCmd {
	return c.GetRedisCmd(ctx).ClusterCountFailureReports(ctx, nodeID)
}

func (c *Redis) ClusterCountKeysInSlot(ctx context.Context, slot int) *redis.IntCmd {
	return c.GetRedisCmd(ctx).ClusterCountKeysInSlot(ctx, slot)
}

func (c *Redis) ClusterDelSlots(ctx context.Context, slots ...int) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterDelSlots(ctx, slots...)
}

func (c *Redis) ClusterDelSlotsRange(ctx context.Context, min, max int) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterDelSlotsRange(ctx, min, max)
}

func (c *Redis) ClusterSaveConfig(ctx context.Context) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterSaveConfig(ctx)
}

func (c *Redis) ClusterSlaves(ctx context.Context, nodeID string) *redis.StringSliceCmd {
	return c.GetRedisCmd(ctx).ClusterSlaves(ctx, nodeID)
}

func (c *Redis) ClusterFailover(ctx context.Context) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterFailover(ctx)
}

func (c *Redis) ClusterAddSlots(ctx context.Context, slots ...int) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterAddSlots(ctx, slots...)
}

func (c *Redis) ClusterAddSlotsRange(ctx context.Context, min, max int) *redis.StatusCmd {
	return c.GetRedisCmd(ctx).ClusterAddSlotsRange(ctx, min, max)
}
