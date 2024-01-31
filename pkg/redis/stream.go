package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func (c *Redis) XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	if a == nil {
		return c.GetRedisCmd(ctx).XAdd(ctx, nil)
	}
	_a := *a
	_a.Stream = c.formatKey(a.Stream)
	return c.GetRedisCmd(ctx).XAdd(ctx, &_a)
}

func (c *Redis) XDel(ctx context.Context, stream string, ids ...string) (int64, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XDel(ctx, stream, ids...).Result()
}

func (c *Redis) XLen(ctx context.Context, stream string) (int64, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XLen(ctx, stream).Result()
}

func (c *Redis) XRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XRange(ctx, stream, start, stop).Result()
}

func (c *Redis) XRangeN(ctx context.Context, stream, start, stop string, count int64) ([]redis.XMessage, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XRangeN(ctx, stream, start, stop, count).Result()
}

func (c *Redis) XRevRange(ctx context.Context, stream string, start, stop string) ([]redis.XMessage, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XRevRange(ctx, stream, start, stop).Result()
}

func (c *Redis) XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) ([]redis.XMessage, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XRevRangeN(ctx, stream, start, stop, count).Result()
}

func (c *Redis) XRead(ctx context.Context, a *redis.XReadArgs) *redis.XStreamSliceCmd {
	if a == nil {
		return c.GetRedisCmd(ctx).XRead(ctx, nil)
	}
	_a := *a
	_a.Streams = c.formatKeys(a.Streams)
	return c.GetRedisCmd(ctx).XRead(ctx, &_a)
}

func (c *Redis) XReadStreams(ctx context.Context, streams ...string) *redis.XStreamSliceCmd {
	_streams := c.formatKeys(streams)
	return c.GetRedisCmd(ctx).XReadStreams(ctx, _streams...)
}

func (c *Redis) XGroupCreate(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XGroupCreate(ctx, stream, group, start)
}

func (c *Redis) XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XGroupCreateMkStream(ctx, stream, group, start)
}

func (c *Redis) XGroupSetID(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XGroupSetID(ctx, stream, group, start)
}

func (c *Redis) XGroupDestroy(ctx context.Context, stream, group string) (int64, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XGroupDestroy(ctx, stream, group).Result()
}

func (c *Redis) XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) (int64, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XGroupCreateConsumer(ctx, stream, group, consumer).Result()
}

func (c *Redis) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) (int64, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XGroupDelConsumer(ctx, stream, group, consumer).Result()
}

func (c *Redis) XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	if a == nil {
		return c.GetRedisCmd(ctx).XReadGroup(ctx, nil)
	}
	_a := *a
	_a.Streams = c.formatKeys(a.Streams)
	return c.GetRedisCmd(ctx).XReadGroup(ctx, &_a)
}

func (c *Redis) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error) {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XAck(ctx, stream, group, ids...).Result()
}

func (c *Redis) XPending(ctx context.Context, stream, group string) *redis.XPendingCmd {
	stream = c.formatKey(stream)
	return c.GetRedisCmd(ctx).XPending(ctx, stream, group)
}

func (c *Redis) XPendingExt(ctx context.Context, a *redis.XPendingExtArgs) *redis.XPendingExtCmd {
	if a == nil {
		return c.GetRedisCmd(ctx).XPendingExt(ctx, nil)
	}
	_a := *a
	_a.Stream = c.formatKey(a.Stream)
	return c.GetRedisCmd(ctx).XPendingExt(ctx, &_a)
}

func (c *Redis) XClaim(ctx context.Context, a *redis.XClaimArgs) ([]redis.XMessage, error) {
	if a == nil {
		return c.GetRedisCmd(ctx).XClaim(ctx, nil).Result()
	}
	_a := *a
	_a.Stream = c.formatKey(a.Stream)
	return c.GetRedisCmd(ctx).XClaim(ctx, &_a).Result()
}

func (c *Redis) XClaimJustID(ctx context.Context, a *redis.XClaimArgs) ([]string, error) {
	if a == nil {
		return c.GetRedisCmd(ctx).XClaimJustID(ctx, nil).Result()
	}
	_a := *a
	_a.Stream = c.formatKey(a.Stream)
	return c.GetRedisCmd(ctx).XClaimJustID(ctx, &_a).Result()
}

func (c *Redis) XAutoClaim(ctx context.Context, a *redis.XAutoClaimArgs) (messages []redis.XMessage, start string, err error) {
	if a == nil {
		return c.GetRedisCmd(ctx).XAutoClaim(ctx, nil).Result()
	}
	_a := *a
	_a.Stream = c.formatKey(a.Stream)
	return c.GetRedisCmd(ctx).XAutoClaim(ctx, &_a).Result()
}

func (c *Redis) XAutoClaimJustID(ctx context.Context, a *redis.XAutoClaimArgs) (ids []string, start string, err error) {
	if a == nil {
		return c.GetRedisCmd(ctx).XAutoClaimJustID(ctx, nil).Result()
	}
	_a := *a
	_a.Stream = c.formatKey(a.Stream)
	return c.GetRedisCmd(ctx).XAutoClaimJustID(ctx, &_a).Result()
}

func (c *Redis) XTrimMaxLen(ctx context.Context, key string, maxLen int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XTrimMaxLen(ctx, key, maxLen).Result()
}

func (c *Redis) XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XTrimMaxLenApprox(ctx, key, maxLen, limit).Result()
}

func (c *Redis) XTrimMinID(ctx context.Context, key string, minID string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XTrimMinID(ctx, key, minID).Result()
}

func (c *Redis) XTrimMinIDApprox(ctx context.Context, key string, minID string, limit int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XTrimMinIDApprox(ctx, key, minID, limit).Result()
}

func (c *Redis) XInfoGroups(ctx context.Context, key string) ([]redis.XInfoGroup, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XInfoGroups(ctx, key).Result()
}

func (c *Redis) XInfoStream(ctx context.Context, key string) (*redis.XInfoStream, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XInfoStream(ctx, key).Result()
}

func (c *Redis) XInfoStreamFull(ctx context.Context, key string, count int) (*redis.XInfoStreamFull, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XInfoStreamFull(ctx, key, count).Result()
}

func (c *Redis) XInfoConsumers(ctx context.Context, key string, group string) ([]redis.XInfoConsumer, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).XInfoConsumers(ctx, key, group).Result()
}
