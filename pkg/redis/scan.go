package redis

import (
	"context"
	"io"

	"github.com/pkg/errors"
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

// doScanFunc 扫描缓存，依次回调匹配的key。
// initialCursor: 初始游标，第一次调用时传0，后续调用传上一次返回的nextCursor。
// scanFunc: 扫描函数，参数为nextCursor，返回值为keys、nextCursor、err。支持 Scan, ScanType, SScan, HScan, ZScan。
// callback: 回调函数，参数为keys，当返回err时，扫描会停止。当返回 io.EOF 时，扫描也会停止，但是 doScanFunc 不会返回错误
// 返回值：
// err: 错误。
func (c *Redis) doScanFunc(ctx context.Context, scanFunc func(ctx context.Context, nextCursor uint64) ([]string, uint64, error), callback func(ctx context.Context, keys []string) error) error {
	var nextCursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, nextCursor, err = scanFunc(ctx, nextCursor)
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err = callback(ctx, keys); err != nil {
				// 如果返回 io.EOF，扫描会停止，但是 doScanFunc 不会返回错误
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
		}

		// 如果nextCursor为0，说明扫描完成
		if nextCursor == 0 {
			break
		}
	}
	return nil
}

// AllScanFunc 扫描缓存，依次回调匹配的key。 参数参考 Scan
func (c *Redis) AllScanFunc(ctx context.Context, pattern string, size int64, callback func(ctx context.Context, keys []string) error) error {
	return c.doScanFunc(ctx, func(ctx context.Context, nextCursor uint64) ([]string, uint64, error) {
		return c.Scan(ctx, pattern, nextCursor, size)
	}, callback)
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

// AllScanTypeFunc 扫描缓存，依次回调匹配的key。 参数参考 ScanType
func (c *Redis) AllScanTypeFunc(ctx context.Context, pattern string, count int64, keyType string, callback func(ctx context.Context, keys []string) error) error {
	return c.doScanFunc(ctx, func(ctx context.Context, nextCursor uint64) ([]string, uint64, error) {
		return c.ScanType(ctx, nextCursor, pattern, count, keyType)
	}, callback)
}

// SScan 扫描set，返回所有匹配的key的members
// https://redis.io/commands/sscan
func (c *Redis) SScan(ctx context.Context, key string, cursor uint64, pattern string, count int64) (members []string, nextCursor uint64, _ error) {
	key = c.formatKey(key) // match是子member，不需要format
	return c.GetRedisCmd(ctx).SScan(ctx, key, cursor, pattern, count).Result()
}

// AllSScanFunc 扫描set，依次回调匹配的key的members。 参数参考 SScan
func (c *Redis) AllSScanFunc(ctx context.Context, key string, pattern string, count int64, callback func(ctx context.Context, members []string) error) error {
	return c.doScanFunc(ctx, func(ctx context.Context, nextCursor uint64) ([]string, uint64, error) {
		return c.SScan(ctx, key, nextCursor, pattern, count)
	}, callback)
}

// HScan 扫描hash，返回所有匹配的key的fields
func (c *Redis) HScan(ctx context.Context, key string, cursor uint64, pattern string, count int64) (field []string, nextCursor uint64, _ error) {
	key = c.formatKey(key) // match是子field，不需要format
	return c.GetRedisCmd(ctx).HScan(ctx, key, cursor, pattern, count).Result()
}

// AllHScanFunc 扫描hash，依次回调匹配的key的fields。 参数参考 HScan
func (c *Redis) AllHScanFunc(ctx context.Context, key string, pattern string, count int64, callback func(ctx context.Context, fields []string) error) error {
	return c.doScanFunc(ctx, func(ctx context.Context, nextCursor uint64) ([]string, uint64, error) {
		return c.HScan(ctx, key, nextCursor, pattern, count)
	}, callback)
}

// ZScan 扫描zset，返回所有匹配的key的members
func (c *Redis) ZScan(ctx context.Context, key string, cursor uint64, pattern string, count int64) (members []string, nextCursor uint64, _ error) {
	key = c.formatKey(key) // match是子member，不需要format
	return c.GetRedisCmd(ctx).ZScan(ctx, key, cursor, pattern, count).Result()
}

// AllZScanFunc 扫描zset，依次回调匹配的key的members。 参数参考 ZScan
func (c *Redis) AllZScanFunc(ctx context.Context, key string, pattern string, count int64, callback func(ctx context.Context, members []string) error) error {
	return c.doScanFunc(ctx, func(ctx context.Context, nextCursor uint64) ([]string, uint64, error) {
		return c.ZScan(ctx, key, nextCursor, pattern, count)
	}, callback)
}
