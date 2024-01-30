package redis

import (
	"context"
)

func (c *Redis) PFAdd(ctx context.Context, key string, elements ...any) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).PFAdd(ctx, key, elements...).Result()
}

func (c *Redis) PFCount(ctx context.Context, keys ...string) (int64, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).PFCount(ctx, _keys...).Result()
}

func (c *Redis) PFMerge(ctx context.Context, dest string, keys ...string) (string, error) {
	dest = c.formatKey(dest)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).PFMerge(ctx, dest, _keys...).Result()
}
