package redis

import (
	"context"
	"time"
)

// Expire 设置过期时间
// EXPIRE key seconds
// https://redis.io/commands/expire
func (c *Redis) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Expire(ctx, key, expiration).Result()
}

// ExpireAt 设置过期时间
// EXPIREAT key timestamp
// https://redis.io/commands/expireat
func (c *Redis) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ExpireAt(ctx, key, tm).Result()
}

// ExpireXX 如果KEY存在，就设置过期时间，否则不设置
// EXPIRE key seconds XX
// https://redis.io/commands/expire
func (c *Redis) ExpireXX(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ExpireXX(ctx, key, expiration).Result()
}

// ExpireNX 如果KEY不存在，就设置过期时间，否则不设置
// EXPIRE key seconds NX
// https://redis.io/commands/expire
func (c *Redis) ExpireNX(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ExpireNX(ctx, key, expiration).Result()
}

// ExpireGT 如果KEY存在，且过期时间大于expiration，就设置过期时间，否则不设置
// EXPIRE key seconds GT
// https://redis.io/commands/expire
func (c *Redis) ExpireGT(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ExpireGT(ctx, key, expiration).Result()
}

// ExpireLT 如果KEY存在，且过期时间小于expiration，就设置过期时间，否则不设置
// EXPIRE key seconds LT
// https://redis.io/commands/expire
func (c *Redis) ExpireLT(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ExpireLT(ctx, key, expiration).Result()
}

// Persist 移除key的过期时间，使key永久有效
// PERSIST key
// https://redis.io/commands/persist
func (c *Redis) Persist(ctx context.Context, key string) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Persist(ctx, key).Result()
}

// PExpire 设置过期时间，单位是毫秒
// PEXPIRE key milliseconds
// https://redis.io/commands/pexpire
func (c *Redis) PExpire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).PExpire(ctx, key, expiration).Result()
}

// PExpireAt 设置过期时间，单位是毫秒
// PEXPIREAT key milliseconds-timestamp
// https://redis.io/commands/pexpireat
func (c *Redis) PExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).PExpireAt(ctx, key, tm).Result()
}

func (c *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).TTL(ctx, key).Result()
}

func (c *Redis) PTTL(ctx context.Context, key string) (time.Duration, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).PTTL(ctx, key).Result()
}
