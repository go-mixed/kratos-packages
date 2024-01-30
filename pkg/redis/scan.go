package redis

import (
	"context"
)

// Scan 扫描缓存，返回所有匹配的key。
// https://redis.io/commands/scan
// pattern: 匹配的key。
// cursor: 游标，第一次调用时传0，后续调用传上一次返回的nextCursor。
// size: 每次扫描的数量，0表示不限制。
// 返回值：
// keys: 扫描到的key。
// nextCursor: 下一次调用时传入的cursor。
// err: 错误。
func (c *Redis) Scan(ctx context.Context, pattern string, cursor uint64, size int64) (keys []string, nextCursor uint64, _ error) {
	pattern = c.formatKey(pattern)
	return c.GetRedisCmd(ctx).Scan(ctx, cursor, pattern, size).Result()
}

// ScanType 扫描缓存，返回所有匹配的key。
// https://redis.io/commands/scan
// cursor: 游标，第一次调用时传0，后续调用传上一次返回的nextCursor。
// pattern: 匹配的key。
// count: 每次扫描的数量，0表示不限制。
// keyType: key的类型，可选值：string、list、set、zset、hash、stream。
func (c *Redis) ScanType(ctx context.Context, cursor uint64, pattern string, count int64, keyType string) (keys []string, nextCursor uint64, _ error) {
	pattern = c.formatKey(pattern)
	return c.GetRedisCmd(ctx).ScanType(ctx, cursor, pattern, count, keyType).Result()
}

// SScan 扫描set，返回所有匹配的key的members
// https://redis.io/commands/sscan
func (c *Redis) SScan(ctx context.Context, key string, cursor uint64, pattern string, count int64) (members []string, nextCursor uint64, _ error) {
	key = c.formatKey(key) // match是子member，不需要format
	return c.GetRedisCmd(ctx).SScan(ctx, key, cursor, pattern, count).Result()
}

// HScan 扫描hash，返回所有匹配的key的fields
func (c *Redis) HScan(ctx context.Context, key string, cursor uint64, pattern string, count int64) (field []string, nextCursor uint64, _ error) {
	key = c.formatKey(key) // match是子field，不需要format
	return c.GetRedisCmd(ctx).HScan(ctx, key, cursor, pattern, count).Result()
}

// ZScan 扫描zset，返回所有匹配的key的members
func (c *Redis) ZScan(ctx context.Context, key string, cursor uint64, pattern string, count int64) (members []string, nextCursor uint64, _ error) {
	key = c.formatKey(key) // match是子member，不需要format
	return c.GetRedisCmd(ctx).ZScan(ctx, key, cursor, pattern, count).Result()
}
