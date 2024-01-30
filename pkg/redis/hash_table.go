package redis

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
)

// HSet 设置哈希表 key 中的字段 field 的值为 value 。
// HSET key field value
// https://redis.io/commands/hset
// 设置成功，返回true
func (c *Redis) HSet(ctx context.Context, key string, member string, value any) (bool, error) {
	key = c.formatKey(key)
	_, err := c.GetRedisCmd(ctx).HSet(ctx, key, member, WrapBinaryMarshaler(value)).Result()
	return err == nil, c.expire(ctx, err, key)
}

// HSetNX 只有在字段 field 不存在时，设置哈希表字段的值。
// HSETNX key field value
// https://redis.io/commands/hsetnx
// 如果key不存在，会返回true，并更新值；否则返回false，不会更新值
func (c *Redis) HSetNX(ctx context.Context, key string, member string, value any) (bool, error) {
	key = c.formatKey(key)

	ok, err := c.GetRedisCmd(ctx).HSetNX(ctx, key, member, WrapBinaryMarshaler(value)).Result()
	// 只有设置成功才设置过期时间
	if ok {
		return ok, c.expire(ctx, err, key)
	}
	return ok, err
}

// HGet 获取哈希表 key 中给定域 field 的值。
// HGET key field
// 如果key不存在，会返回false
func (c *Redis) HGet(ctx context.Context, key string, field string, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).HGet(ctx, key, field)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// HGetAll 获取在哈希表中指定 key 的所有字段和值
// HGETALL key
// https://redis.io/commands/hgetall
// actual的结构是&struct{A int `redis:"a"`, ...}，并且无法内嵌Struct、Map，功能十分有限，如果需要更多功能，可以使用ModernCache.HGetAll
// 支持结构详见：github.com/redis/go-redis/v9/internal/hscan/hscan.go的decoders
func (c *Redis) HGetAll(ctx context.Context, key string, actual any) (map[string]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).HGetAll(ctx, key)

	err := utils.IfFunc(actual != nil, func() error { return res.Scan(actual) }, func() error { return res.Err() })
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// HDel 删除一个或多个哈希表字段
// HDEL key field [field ...]
// https://redis.io/commands/hdel
// 返回值：
// 1. 被成功删除字段的数量，不包括被忽略的字段。
func (c *Redis) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	key = c.formatKey(key) // fields是子field，不需要format
	return c.GetRedisCmd(ctx).HDel(ctx, key, fields...).Result()
}

// HExists 查看哈希表 key 中，给定域 field 是否存在。
// HEXISTS key field
// https://redis.io/commands/hexists
func (c *Redis) HExists(ctx context.Context, key, field string) (ok bool, _ error) {
	key = c.formatKey(key) // field是子field，不需要format
	return c.GetRedisCmd(ctx).HExists(ctx, key, field).Result()
}

// HIncrBy 为哈希表 key 中的指定字段的整数值加上增量 increment 。
// HINCRBY key field increment
// https://redis.io/commands/hincrby
func (c *Redis) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	key = c.formatKey(key) // field是子field，不需要format
	return c.GetRedisCmd(ctx).HIncrBy(ctx, key, field, incr).Result()
}

// HIncrByFloat 为哈希表 key 中的指定字段的浮点数值加上增量 increment 。
// HINCRBYFLOAT key field increment
// https://redis.io/commands/hincrbyfloat
func (c *Redis) HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error) {
	key = c.formatKey(key) // field是子field，不需要format
	return c.GetRedisCmd(ctx).HIncrByFloat(ctx, key, field, incr).Result()
}

// HKeys 获取所有哈希表中的字段
// HKEYS key
// https://redis.io/commands/hkeys
func (c *Redis) HKeys(ctx context.Context, key string) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).HKeys(ctx, key).Result()
}

// HLen 获取哈希表中字段的数量
// HLEN key
// https://redis.io/commands/hlen
func (c *Redis) HLen(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).HLen(ctx, key).Result()
}

// HMGet 获取所有给定字段的值
// HMGET key field [field ...]
// https://redis.io/commands/hmget
// actual的结构为&Struct{A int `redis:"a"`, ...}，并且无法内嵌Struct、Map，功能十分有限，如果需要更多功能，可以使用ModernCache.HMGet
func (c *Redis) HMGet(ctx context.Context, key string, actual any, fields ...string) ([]string, error) {
	key = c.formatKey(key) // fields是子field，不需要format
	res := c.GetRedisCmd(ctx).HMGet(ctx, key, fields...)

	err := utils.IfFunc(actual != nil, func() error { return res.Scan(actual) }, func() error { return res.Err() })
	if err != nil {
		return nil, filterNil(err)
	}
	return utils.InterfacesToStrings(res.Val()), nil
}

// HMSet 同时将多个 field-value (域-值)对设置到哈希表 key 中。
// HMSET key field value [field value ...]
// https://redis.io/commands/hmset
// 设置成功，返回true
func (c *Redis) HMSet(ctx context.Context, key string, kvs map[string]any) (bool, error) {
	key = c.formatKey(key) // values中包含子field，不需要format
	return c.GetRedisCmd(ctx).HMSet(ctx, key, WrapMapBinaryMarshaler(kvs)).Result()
}

// HVals 获取哈希表中所有值
// HVALS key
// https://redis.io/commands/hvals
func (c *Redis) HVals(ctx context.Context, key string, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).HVals(ctx, key)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// HRandField 从哈希表中随机获取一个字段，可以指定获取多个。
// HRandField key [count [WITHVALUES]]
// https://redis.io/commands/hscan
// 如果withValues为true，则同时返回值：[]string{field, value, field, value, ...}，不然只返回[]string{field, field, ...}
func (c *Redis) HRandField(ctx context.Context, key string, count int, withValues bool) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).HRandField(ctx, key, count).Result()
}
