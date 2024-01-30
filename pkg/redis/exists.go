package redis

import (
	"context"
)

// Exists 判断key是否存在，如果传递多个key，则返回存在的key的数量（重复的key会算多次）
// EXISTS key [key ...]
// https://redis.io/commands/exists
// 返回值：
// 1. int64: 存在的key的数量
// 2. error: 失败时返回的错误
func (c *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).Exists(ctx, _keys...).Result()
}

// Has 判断key是否存在
// EXISTS key
// https://redis.io/commands/exists
// 返回值：
// 1. bool: key是否存在
// 2. error: 失败时返回的错误
func (c *Redis) Has(ctx context.Context, key string) (bool, error) {
	key = c.formatKey(key)
	return c.HasAny(ctx, key)
}

// HasAny 判断任意key是否存在
// EXISTS key [key ...]
// https://redis.io/commands/exists
// 返回值：
// 1. bool: 任意key是否存在
// 2. error: 失败时返回的错误
func (c *Redis) HasAny(ctx context.Context, keys ...string) (bool, error) {
	_keys := c.formatKeys(keys)
	count, err := c.GetRedisCmd(ctx).Exists(ctx, _keys...).Result()
	return count > 0, err
}

// HaveAll 判断所有key是否存在
// EXISTS key [key ...]
// https://redis.io/commands/exists
// 返回值：
// 1. bool: 所有key是否存在
// 2. error: 失败时返回的错误
func (c *Redis) HaveAll(ctx context.Context, keys ...string) (bool, error) {
	_keys := c.formatKeys(keys)
	count, err := c.GetRedisCmd(ctx).Exists(ctx, _keys...).Result()
	return count == int64(len(_keys)), err
}
