package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// GetBit 返回 key 中第 offset 位的值，0 或 1
// GETBIT key offset
// https://redis.io/commands/getbit
func (c *Redis) GetBit(ctx context.Context, key string, offset int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GetBit(ctx, key, offset).Result()
}

// SetBit 设置 key 中第 offset 位的值，0 或 1
// SETBIT key offset value
// https://redis.io/commands/setbit
func (c *Redis) SetBit(ctx context.Context, key string, offset int64, value int) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SetBit(ctx, key, offset, value).Result()
}

// BitCount 统计 key 中Start~End的二进制位为 1 的数量
// BITCOUNT key [start end [BYTE | BIT]]
// https://redis.io/commands/bitcount
func (c *Redis) BitCount(ctx context.Context, key string, bitCount *redis.BitCount) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).BitCount(ctx, key, bitCount).Result()
}

// BitOpAnd 将多个 key 的值，逻辑与（AND）后，存储到 destKey
// BITOP AND destKey key [key ...]
// https://redis.io/commands/bitop
func (c *Redis) BitOpAnd(ctx context.Context, destKey string, keys ...string) (int64, error) {
	destKey = c.formatKey(destKey)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).BitOpAnd(ctx, destKey, _keys...).Result()
}

// BitOpOr 将多个 key 的值，逻辑或（OR）后，存储到 destKey
// BITOP OR destKey key [key ...]
// https://redis.io/commands/bitop
func (c *Redis) BitOpOr(ctx context.Context, destKey string, keys ...string) (int64, error) {
	destKey = c.formatKey(destKey)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).BitOpOr(ctx, destKey, _keys...).Result()
}

// BitOpXor 将多个 key 的值，逻辑异或（XOR）后，存储到 destKey
// BITOP XOR destKey key [key ...]
// https://redis.io/commands/bitop
func (c *Redis) BitOpXor(ctx context.Context, destKey string, keys ...string) (int64, error) {
	destKey = c.formatKey(destKey)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).BitOpXor(ctx, destKey, _keys...).Result()
}

// BitOpNot 将 key 的值，逻辑非（NOT）后，存储到 destKey
// BITOP NOT destKey key
// https://redis.io/commands/bitop
func (c *Redis) BitOpNot(ctx context.Context, destKey string, key string) (int64, error) {
	destKey = c.formatKey(destKey)
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).BitOpNot(ctx, destKey, key).Result()
}

// BitPos 返回 key 中第一个值为 bit 的索引，bit 为 1 或 0
// BITPOS key bit [start [end [BYTE | BIT]]]
// https://redis.io/commands/bitpos
func (c *Redis) BitPos(ctx context.Context, key string, bit int64, pos ...int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).BitPos(ctx, key, bit, pos...).Result()
}

// BitField 对 key 中的二进制位进行操作
// https://redis.io/commands/bitfield
func (c *Redis) BitField(ctx context.Context, key string, args ...any) ([]int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).BitField(ctx, key, args...).Result()
}
