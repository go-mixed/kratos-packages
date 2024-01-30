package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"time"
)

// BLPop 阻塞直到有元素可弹出或超时，从左侧弹出
// BLPOP key [key ...] timeout
// https://redis.io/commands/blpop
// 参数：
// 1. timeout: 超时时间
// 2. actual: 弹出的元素通过反射赋值给actual，actual的类型为&[]struct{...}
// 返回值：
// 1. []string: 弹出的元素
// 2. error: 失败时返回的错误
func (c *Redis) BLPop(ctx context.Context, timeout time.Duration, actual any, keys ...string) ([]string, error) {
	_keys := c.formatKeys(keys)
	res := c.GetRedisCmd(ctx).BLPop(ctx, timeout, _keys...)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// BRPop 阻塞直到有元素可弹出或超时，从右侧弹出
// BRPOP key [key ...] timeout
// https://redis.io/commands/brpop
// 参数：
// 1. timeout: 超时时间
// 2. actual: 弹出的元素通过反射赋值给actual，actual的类型为&[]struct{...}
// 返回值：
// 1. []string: 弹出的元素
// 2. error: 失败时返回的错误
func (c *Redis) BRPop(ctx context.Context, timeout time.Duration, actual any, keys ...string) ([]string, error) {
	_keys := c.formatKeys(keys)
	res := c.GetRedisCmd(ctx).BRPop(ctx, timeout, _keys...)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// BRPopLPush 阻塞直到有元素可弹出或超时，从右侧弹出，然后从左侧插入
// BRPOPLPUSH source destination timeout
// https://redis.io/commands/brpoplpush
// 参数：
// 1. timeout: 超时时间
// 2. actual: 弹出的元素通过反射赋值给actual
// 返回值：
// 1. string: 弹出的元素
// 2. error: 失败时返回的错误
func (c *Redis) BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration, actual any) (string, error) {
	source = c.formatKey(source)
	destination = c.formatKey(destination)
	res := c.GetRedisCmd(ctx).BRPopLPush(ctx, source, destination, timeout)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// LIndex 获取列表指定下标的元素
// LINDEX key index
// https://redis.io/commands/lindex
// 参数：
// 1. index: 下标，从0开始
// 2. actual: 获取的元素通过反射赋值给actual
// 返回值：
// 1. string: 列表中下标为index的元素
func (c *Redis) LIndex(ctx context.Context, key string, index int64, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).LIndex(ctx, key, index)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// LInsert 在列表的元素前或后插入元素
// LINSERT key BEFORE|AFTER pivot element
// https://redis.io/commands/linsert
// 参数：
// 1. op: BEFORE|AFTER
// 2. pivot: 列表中的元素
// 3. value: 要插入的元素
// 返回值：
// 1. int64: 插入后列表的长度
// 2. error: 失败时返回的错误
func (c *Redis) LInsert(ctx context.Context, key, op string, pivot, value any) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LInsert(ctx, key, op, WrapBinaryMarshaler(pivot), WrapBinaryMarshaler(value)).Result()
}

// LInsertBefore 在列表的元素前插入元素
// LINSERT key BEFORE pivot element
// https://redis.io/commands/linsert
// 参数：
// 1. pivot: 列表中的元素
// 2. value: 要插入的元素
// 返回值：
// 1. int64: 插入后列表的长度
// 2. error: 失败时返回的错误
func (c *Redis) LInsertBefore(ctx context.Context, key string, pivot, value any) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LInsertBefore(ctx, key, WrapBinaryMarshaler(pivot), WrapBinaryMarshaler(value)).Result()
}

// LInsertAfter 在列表的元素后插入元素
// LINSERT key AFTER pivot element
// https://redis.io/commands/linsert
// 参数：
// 1. pivot: 列表中的元素
// 2. value: 要插入的元素
// 返回值：
// 1. int64: 插入后列表的长度
// 2. error: 失败时返回的错误
func (c *Redis) LInsertAfter(ctx context.Context, key string, pivot, value any) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LInsertAfter(ctx, key, WrapBinaryMarshaler(pivot), WrapBinaryMarshaler(value)).Result()
}

// LLen 获取列表的长度
// LLEN key
// https://redis.io/commands/llen
// 返回值：
// 1. int64: 列表的长度
func (c *Redis) LLen(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LLen(ctx, key).Result()
}

// LPop 移除并返回列表的第一个元素（左侧）
// LPOP key
// https://redis.io/commands/lpop
// 参数：
// 1. actual: 移除的元素通过反射赋值给actual
// 返回值：
// 1. string: 移除的元素
// 2. error: 失败时返回的错误
func (c *Redis) LPop(ctx context.Context, key string, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).LPop(ctx, key)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// LPopCount 移除并返回列表的count个元素（左侧）
// LPOP key	count
// https://redis.io/commands/lpop
// 参数：
// 1. count: 移除的元素个数
// 2. actual: 移除的元素通过反射赋值给actual，actual的类型为&[]struct{...}
// 返回值：
// 1. []string: 移除的元素
// 2. error: 失败时返回的错误
func (c *Redis) LPopCount(ctx context.Context, key string, count int, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).LPopCount(ctx, key, count)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// LPos 获取列表中指定元素的下标
// LPOS key element [RANK rank] [COUNT num-matches] [MAXLEN len]
// https://redis.io/commands/lpos
func (c *Redis) LPos(ctx context.Context, key string, value string, args redis.LPosArgs) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LPos(ctx, key, value, args).Result()
}

// LPosCount 获取列表中指定元素的下标
// LPOS key element [RANK rank] [COUNT num-matches] [MAXLEN len]
// https://redis.io/commands/lpos
func (c *Redis) LPosCount(ctx context.Context, key string, value string, count int64, args redis.LPosArgs) ([]int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LPosCount(ctx, key, value, count, args).Result()
}

// LPush 将一个或多个值插入到列表头部
// LPUSH key value [value ...]
// https://redis.io/commands/lpush
// 返回值：
// 1. int64: 列表的长度
// 2. error: 失败时返回的错误
func (c *Redis) LPush(ctx context.Context, key string, values ...any) (int64, error) {
	key = c.formatKey(key)
	_vals := lo.Map(values, func(v any, _ int) any { return WrapBinaryMarshaler(v) })
	return c.GetRedisCmd(ctx).LPush(ctx, key, _vals).Result()
}

// LPushX 将一个值插入到列表头部，如果列表不存在，则不执行任何操作
// LPUSHX key value
// https://redis.io/commands/lpushx
// 返回值：
// 1. int64: 列表的长度
// 2. error: 失败时返回的错误
func (c *Redis) LPushX(ctx context.Context, key string, values ...any) (int64, error) {
	key = c.formatKey(key)
	_vals := lo.Map(values, func(v any, _ int) any { return WrapBinaryMarshaler(v) })
	return c.GetRedisCmd(ctx).LPushX(ctx, key, _vals...).Result()
}

// LRange 从开头（左侧）获取列表指定范围内的元素
// LRANGE key start stop
// https://redis.io/commands/lrange
// 参数：
// 1. start: 开始下标，从0开始
// 2. stop: 结束下标，-1表示最后一个元素
// 3. actual: 获取的元素通过反射赋值给actual，actual的类型为&[]struct{...}
// 返回值：
// 1. []string: 列表中指定范围内的元素
// 2. error: 失败时返回的错误
func (c *Redis) LRange(ctx context.Context, key string, start, stop int64, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).LRange(ctx, key, start, stop)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// LRem 从开头（左侧）移除count个列表元素 Rem = Remove
// LREM key count value
// https://redis.io/commands/lrem
// 参数：
// 1. count: 要移除的元素个数
// 2. value: 要移除的元素
// 返回值：
// 1. int64: 被移除的元素个数
// 2. error: 失败时返回的错误
func (c *Redis) LRem(ctx context.Context, key string, count int64, value any) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LRem(ctx, key, count, WrapBinaryMarshaler(value)).Result()
}

// LSet 设置列表指定下标的元素（从左侧0开始）
// LSET key index value
// https://redis.io/commands/lset
// 参数：
// 1. index: 下标，从0开始
// 返回值：
// 1. string: OK
// 2. error: 失败时返回的错误
func (c *Redis) LSet(ctx context.Context, key string, index int64, value any) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LSet(ctx, key, index, WrapBinaryMarshaler(value)).Result()
}

// LTrim 保留列表指定范围内的元素（从左侧0开始）
// LTRIM key start stop
// https://redis.io/commands/ltrim
// 参数：
// 1. start: 开始下标，从0开始
// 2. stop: 结束下标，-1表示最后一个元素
// 返回值：
// 1. string: OK
// 2. error: 失败时返回的错误
func (c *Redis) LTrim(ctx context.Context, key string, start, stop int64) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).LTrim(ctx, key, start, stop).Result()
}

// RPop 移除并返回列表的最后一个元素（右侧）
// RPOP key
// https://redis.io/commands/rpop
// 参数：
// 1. actual: 移除的元素通过反射赋值给actual
// 返回值：
// 1. string: 移除的元素
func (c *Redis) RPop(ctx context.Context, key string, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).RPop(ctx, key)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// RPopCount 移除并返回列表的count个元素（右侧）
// RPOP key	count
// https://redis.io/commands/rpop
// 参数：
// 1. count: 移除的元素个数
// 2. actual: 移除的元素通过反射赋值给actual，actual的类型为&[]struct{...}
// 返回值：
// 1. []string: 移除的元素
// 2. error: 失败时返回的错误
func (c *Redis) RPopCount(ctx context.Context, key string, count int, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).RPopCount(ctx, key, count)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// RPopLPush 移除列表的最后一个元素（右侧），并将该元素添加到另一个列表（左侧）
// RPOPLPUSH source destination
// https://redis.io/commands/rpoplpush
// 参数：
// 1. actual: 弹出的元素通过反射赋值给actual
// 返回值：
// 1. bool: 是否成功
// 2. error: 失败时返回的错误
func (c *Redis) RPopLPush(ctx context.Context, source, destination string, actual any) (bool, error) {
	source = c.formatKey(source)
	destination = c.formatKey(destination)
	res := c.GetRedisCmd(ctx).RPopLPush(ctx, source, destination)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return false, filterNil(err)
	}
	return true, nil
}

// RPush 将一个或多个值插入到列表尾部（右侧）
// RPUSH key value [value ...]
// https://redis.io/commands/rpush
func (c *Redis) RPush(ctx context.Context, key string, values ...any) (int64, error) {
	key = c.formatKey(key)
	_vals := lo.Map(values, func(v any, _ int) any { return WrapBinaryMarshaler(v) })
	return c.GetRedisCmd(ctx).RPush(ctx, key, _vals...).Result()
}

// RPushX 将一个值插入到列表尾部（右侧），如果列表不存在，则不执行任何操作
// RPUSHX key value
// https://redis.io/commands/rpushx
func (c *Redis) RPushX(ctx context.Context, key string, values ...any) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).RPushX(ctx, key, values...).Result()
}

// LMove 将列表source中的srcpos元素弹出，并插入到列表destination的destpos位置
// LMOVE source destination LEFT|RIGHT LEFT|RIGHT
// https://redis.io/commands/lmove
// 参数：
// 1. srcpos: LEFT|RIGHT
// 2. destpos: LEFT|RIGHT
// 3. actual: 弹出的元素通过反射赋值给actual
// 返回值：
// 1. bool: 是否成功
// 2. error: 失败时返回的错误
func (c *Redis) LMove(ctx context.Context, source, destination, srcpos, destpos string, actual any) (bool, error) {
	source = c.formatKey(source)
	destination = c.formatKey(destination)
	res := c.GetRedisCmd(ctx).LMove(ctx, source, destination, srcpos, destpos)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return false, filterNil(err)
	}
	return true, nil
}

// BLMove 阻塞直到有元素可弹出或超时，从srcpos弹出，然后插入到另一个列表的destpos位置
// BLMOVE source destination LEFT|RIGHT LEFT|RIGHT timeout
// https://redis.io/commands/blmove
// 参数：
// 1. srcpos: LEFT|RIGHT
// 2. destpos: LEFT|RIGHT
// 3. timeout: 超时时间
// 4. actual: 弹出的元素通过反射赋值给actual
// 返回值：
// 1. bool: 是否成功
// 2. error: 失败时返回的错误
func (c *Redis) BLMove(ctx context.Context, source, destination, srcpos, destpos string, timeout time.Duration, actual any) (bool, error) {
	source = c.formatKey(source)
	destination = c.formatKey(destination)
	res := c.GetRedisCmd(ctx).BLMove(ctx, source, destination, srcpos, destpos, timeout)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return false, filterNil(err)
	}
	return true, nil
}
