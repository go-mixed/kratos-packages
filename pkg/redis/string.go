package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"time"
)

// Remember 如果有缓存，则反射并设置actual的值；不然就执行callback(act)，并在callback中设置actual的值；如果callback没有报错则将actual设置缓存。
// 如果你希望调用callback之后不被设置缓存，需要返回ErrSkipCache，本函数会检查这个特殊的异常，并跳过设置缓存，然后error返回nil。
func (c *Redis) Remember(ctx context.Context, key string, actual any, callback func(ctx context.Context, actual any) error) error {
	res, err := c.Get(ctx, key, actual)
	if err != nil {
		return err
	} else if res != "" { // 存在于缓存中，直接返回
		return nil
	}

	if callback != nil {
		if err = callback(ctx, actual); err != nil {
			return err
		}
		// 不保存空值
		if !c.options.SaveEmptyOnRemember && utils.IsZero(actual) {
			return nil
		}
		return c.Set(ctx, key, actual)
	}

	return nil
}

// Get 获取缓存，需要传递actual来赋值。
// GET key
// https://redis.io/commands/get
// 参数：
// 1. actual: 将缓存的值反射到actual中，任意类型。为nil时不反射。
// 返回值：
// 1. string: 缓存的值，为空可能是因为key不存在
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) Get(ctx context.Context, key string, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).Get(ctx, key)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// GetRange 返回key对应的字符串value的子串，子串的起始位置为start（包含），结束位置为end（包含）。
// GETRANGE key start end
// https://redis.io/commands/getrange
// 参数：
// 1. start: 起始位置，包含
// 2. end: 结束位置，包含。如果为负数，则表示倒数第几个字符
// 3. actual: 将缓存的值反射到actual中，任意类型。为nil时不反射。
// 返回值：
// 1. string: 缓存的值（start, end)，为空可能是因为key不存在
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) GetRange(ctx context.Context, key string, start, end int64, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).GetRange(ctx, key, start, end)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// MGet 批量获取缓存，需要传递actuals来赋值。
// MGET key [key ...]
// https://redis.io/commands/mget
// actual的结构是&[]*struct{A int `redis:"a"`, ...}，并且无法内嵌Struct、Map，功能十分有限，如果需要更多功能，可以使用ModernCache.MGet
func (c *Redis) MGet(ctx context.Context, keys []string, actual any) ([]string, error) {
	_keys := c.formatKeys(keys)
	res := c.GetRedisCmd(ctx).MGet(ctx, _keys...)

	err := utils.IfFunc(actual != nil, func() error { return res.Scan(actual) }, func() error { return res.Err() })
	if err != nil {
		return nil, filterNil(err)
	}

	return interfacesToStrings(res.Val()), nil
}

// Set 设置缓存，value为任意对象。
// SET key value [expiration]
// https://redis.io/commands/set
func (c *Redis) Set(ctx context.Context, key string, value any) error {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Set(ctx, key, WrapBinaryMarshaler(value), c.options.Expiration).Err()
}

// MSet 批量设置缓存，value为任意对象。
// MSET key value [key value ...]
// https://redis.io/commands/mset
func (c *Redis) MSet(ctx context.Context, kvs map[string]any) (string, error) {
	_m := c.formatMKeys(kvs)
	status, err := c.GetRedisCmd(ctx).MSet(ctx, _m).Result()
	return status, c.expire(ctx, err, lo.Keys(kvs)...)
}

// MSetNX 批量设置缓存，value为任意对象，只有在key不存在时才设置。
// MSETNX key value [key value ...]
// https://redis.io/commands/msetnx
func (c *Redis) MSetNX(ctx context.Context, kvs map[string]any) (bool, error) {
	_m := c.formatMKeys(kvs)
	ok, err := c.GetRedisCmd(ctx).MSetNX(ctx, _m).Result()
	if ok { // 设置成功了，设置过期时间
		return ok, c.expire(ctx, err, lo.Keys(kvs)...)
	}
	return ok, err
}

// StrLen 返回key对应的字符串value的长度。
// STRLEN key
// https://redis.io/commands/strlen
func (c *Redis) StrLen(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).StrLen(ctx, key).Result()
}

// SetEx 设置缓存，value为任意对象，同时设置过期时间。
// SETEX key seconds value [NX|XX]
// https://redis.io/commands/setex
func (c *Redis) SetEx(ctx context.Context, key string, value any, expiration time.Duration) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SetEx(ctx, key, value, expiration).Result()
}

// SetNX 如果KEY不存在，就设置缓存，value为任意对象。注意：这是原子性的
// SETNX key value [expiration] NX
// https://redis.io/commands/setnx
// 设置成功返回true（即key不存在）。
func (c *Redis) SetNX(ctx context.Context, key string, value any) (ok bool, _ error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SetNX(ctx, key, WrapBinaryMarshaler(value), c.options.Expiration).Result()
}

// SetXX 如果KEY存在，就设置缓存，value为任意对象。注意：这是原子性的
// SETNX key value [expiration] XX
// https://redis.io/commands/setnx
// 设置成功返回true（即key存在）。
func (c *Redis) SetXX(ctx context.Context, key string, value any) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SetXX(ctx, key, WrapBinaryMarshaler(value), c.options.Expiration).Result()
}

// SetArgs 传入SET的参数，可以设置NX、XX、EX、PX等参数
// SET key value [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
// https://redis.io/commands/set
// 如果不满足XX、NX的条件，会返回redis.Nil的错误：https://redis.io/commands/set/#return
func (c *Redis) SetArgs(ctx context.Context, key string, value any, a redis.SetArgs) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SetArgs(ctx, key, value, a).Result()
}

// SetRange 将key对应的字符串value的子串从指定的offset处开始，替换为value。
// SETRANGE key offset value
// https://redis.io/commands/setrange
// 参数：
// 1. offset: 偏移量，从0开始
// 2. value: 替换的值
// 返回值：
// 1. int64: 替换后的字符串长度
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) SetRange(ctx context.Context, key string, offset int64, value string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SetRange(ctx, key, offset, value).Result()
}

// Incr 将key对应的数字加1，如果key不存在，则设置为0后再加1。
// INCR key
// https://redis.io/commands/incr
// 返回值：
// 1. int64: 加1后的值
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) Incr(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Incr(ctx, key).Result()
}

// IncrBy 将key对应的数字加value，如果key不存在，则设置为0后再加value。
// INCRBY key increment
// https://redis.io/commands/incrby
// 返回值：
// 1. int64: 加value后的值
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).IncrBy(ctx, key, value).Result()
}

// IncrByFloat 将key对应的数字加value，如果key不存在，则设置为0后再加value。
// INCRBYFLOAT key increment
// https://redis.io/commands/incrbyfloat
// 返回值：
// 1. float64: 加value后的值
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).IncrByFloat(ctx, key, value).Result()
}

// Decr 将key对应的数字减1，如果key不存在，则设置为0后再减1。
// DECR key
// https://redis.io/commands/decr
// 返回值：
// 1. int64: 减1后的值
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) Decr(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Decr(ctx, key).Result()
}

// DecrBy 将key对应的数字减value，如果key不存在，则设置为0后再减value。
// DECRBY key decrement
// https://redis.io/commands/decrby
// 返回值：
// 1. int64: 减value后的值
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).DecrBy(ctx, key, value).Result()
}

// Append 将value追加到key对应的字符串的末尾。
// APPEND key value
// https://redis.io/commands/append
// 返回值：
// 1. int64: 追加后的字符串长度
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) Append(ctx context.Context, key, value string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Append(ctx, key, value).Result()
}

// GetSet 将key对应的字符串设置为value，并返回key对应的旧值。
// GETSET key value
// https://redis.io/commands/getset
// 参数：
// 1. actual: 将缓存的值反射到actual中，任意类型。为nil时不反射。
// 返回值：
// 1. string: 旧值，为空可能是因为key不存在
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) GetSet(ctx context.Context, key string, value any, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).GetSet(ctx, key, WrapBinaryMarshaler(value))

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// GetEx 获取缓存，需要传递actual来赋值，同时设置过期时间。
// GETEX key [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | PERSIST]
// https://redis.io/commands/getex
// 参数：
// 1. expiration: 待设置的过期时间
// 2. actual: 将缓存的值反射到actual中，任意类型。为nil时不反射。
// 返回值：
// 1. string: 缓存的值，为空可能是因为key不存在
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) GetEx(ctx context.Context, key string, expiration time.Duration, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).GetEx(ctx, key, expiration)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}

// GetDel 获取缓存，需要传递actual来赋值，同时删除缓存。
// GETDEL key
// https://redis.io/commands/getdel
// 参数：
// 1. actual: 将缓存的值反射到actual中，任意类型。为nil时不反射。
// 返回值：
// 1. string: 缓存的值，为空可能是因为key不存在
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) GetDel(ctx context.Context, key string, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).GetDel(ctx, key)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), nil
}
