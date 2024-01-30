package redis

import (
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"time"
)

var ErrInvalidLength = errors.New("keys and values length not match")

type ModernRedis[T any] struct {
	*Redis
}

func NewModernRedis[T any](redis *Redis) *ModernRedis[T] {
	return &ModernRedis[T]{
		Redis: redis,
	}
}

func AsModernRedis[T any, N any](src *ModernRedis[T]) *ModernRedis[N] {
	return &ModernRedis[N]{
		Redis: src.Redis,
	}
}

func (c *ModernRedis[T]) makeT(val string) (T, error) {
	if val == "" {
		var nilT T
		return nilT, nil
	}

	var t T = utils.New[T]()

	err := utils.IfFunc(utils.IsPtr(t), func() error {
		return Scan(val, t)
	}, func() error {
		return Scan(val, &t)
	})

	return t, err
}

// makeMap 将keys和values转化为map[string]T
func (c *ModernRedis[T]) makeMap(keys []string, values []string) (map[string]T, error) {
	if len(keys) != len(values) {
		return nil, ErrInvalidLength
	}

	m := make(map[string]T)
	for i, key := range keys {
		var t T = utils.New[T]()

		err := utils.IfFunc(utils.IsPtr(t), func() error {
			return Scan(values[i], t)
		}, func() error {
			return Scan(values[i], &t)
		})
		if err != nil {
			return nil, err
		}
		m[key] = t
	}
	return m, nil
}

// makeSlice 将values转化为[]T
func (c *ModernRedis[T]) makeSlice(values []string) ([]T, error) {
	var ts []T
	for _, s := range values {
		var t T = utils.New[T]()
		err := utils.IfFunc(utils.IsPtr(t), func() error {
			return Scan(s, t)
		}, func() error {
			return Scan(s, &t)
		})
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

// Remember 先尝试获取缓存，如果没有获取到，则调用callback()获取值，并设置缓存。
// 当SaveEmptyOnRemember为true时，即使callback()返回空值（空字符串、0、0长度的map或slice、以及均为空值的struct），也会设置缓存。
func (c *ModernRedis[T]) Remember(ctx context.Context, key string, callback func(ctx context.Context) (T, error),
) (T, error) {
	var err error
	var ok bool
	var actual T
	var nilT T

	ok, actual, err = c.Get(ctx, key)
	if err != nil {
		return nilT, err
	} else if ok { // 存在于缓存中，直接返回
		// get不为空值，并且没有错误。但是actual没有获取到值，说明出现了未知的Scan错误
		// 仅仅检查指针类型==nil
		if utils.IsNil(actual) {
			return nilT, errors.New("redis: get empty actual of " + key)
		}
		return actual, nil
	}

	if callback != nil {
		if actual, err = callback(ctx); err != nil { // 从callback中获取值
			return nilT, err
		}
		// 不保存空值
		if !c.Redis.options.SaveEmptyOnRemember && utils.IsZero(actual) {
			return nilT, nil
		}
		// 设置缓存
		if err = c.Redis.Set(ctx, key, actual); err != nil {
			return nilT, err
		}

		return actual, nil
	}

	panic("redis: callback of remember is nil")
}

// Get 获取缓存
// https://redis.io/commands/get
// 返回值：是否存在，value转为T类型，error
func (c *ModernRedis[T]) Get(ctx context.Context, key string) (bool, T, error) {
	res, err := c.Redis.Get(ctx, key, nil)
	if err != nil {
		var nilT T
		return false, nilT, err
	}
	t, err := c.makeT(res)
	return res != "", t, err
}

// MGet 批量获取缓存
// https://redis.io/commands/mget
// 返回值：key -> T
func (c *ModernRedis[T]) MGet(ctx context.Context, keys ...string) (map[string]T, error) {
	mstr, err := c.Redis.MGet(ctx, keys, nil)
	if err != nil {
		return nil, err
	}

	return c.makeMap(keys, mstr)
}

// HGet 返回哈希表 key 中给定域 field 的值。
// https://redis.io/commands/hget
// 如果key、field不存在，会返回false
// value转为T类型
func (c *ModernRedis[T]) HGet(ctx context.Context, key string, field string) (bool, T, error) {
	res, err := c.Redis.HGet(ctx, key, field, nil)
	if err != nil {
		var nilT T
		return false, nilT, err
	}
	t, err := c.makeT(res)
	return res != "", t, err
}

// HMGet 返回哈希表 key 中，一个或多个给定field的值。
// https://redis.io/commands/hmget
// 返回值：field -> key, value -> T
func (c *ModernRedis[T]) HMGet(ctx context.Context, key string, fields ...string) (map[string]T, error) {
	mstr, err := c.Redis.HMGet(ctx, key, nil, fields...)
	if err != nil {
		return nil, err
	}

	return c.makeMap(fields, mstr)
}

// HGetAll 返回哈希表 key 中，所有的field和值。
// https://redis.io/commands/hgetall
// 返回值：field -> key, value -> T
func (c *ModernRedis[T]) HGetAll(ctx context.Context, key string) (map[string]T, error) {
	mstr, err := c.Redis.HGetAll(ctx, key, nil)
	if err != nil {
		return nil, err
	}

	return c.makeMap(lo.Keys(mstr), lo.Values(mstr))
}

// ZRange 返回有序集 key 中，指定区间内的成员。
// 其中成员的位置按 score 值递增(从小到大)来排列。
// 具有相同 score 值的成员按字典序(lexicographical order )来排列。
// https://redis.io/commands/zrange
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZRange(ctx context.Context, key string, start int64, stop int64) ([]T, error) {
	mstr, err := c.Redis.ZRange(ctx, key, start, stop)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// ZRangeByScore
// https://redis.io/commands/zrangebyscore
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]T, error) {
	mstr, err := c.Redis.ZRangeByScore(ctx, key, opt)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// ZRangeByLex
// https://redis.io/commands/zrangebylex
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) ([]T, error) {
	mstr, err := c.Redis.ZRangeByLex(ctx, key, opt)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// ZRangeArgs
// https://redis.io/commands/zrange
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZRangeArgs(ctx context.Context, args redis.ZRangeArgs) ([]T, error) {
	mstr, err := c.Redis.ZRangeArgs(ctx, args)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// ZRevRange 返回有序集 key 中，指定区间内的成员。
// 其中成员的位置按 score 值递减(从大到小)来排列。
// 具有相同 score 值的成员按字典序的逆序(reverse lexicographical order)排列。
// https://redis.io/commands/zrevrange
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZRevRange(ctx context.Context, key string, start int64, stop int64) ([]T, error) {
	mstr, err := c.Redis.ZRevRange(ctx, key, start, stop)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// ZRevRangeByScore
// https://redis.io/commands/zrevrangebyscore
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]T, error) {
	mstr, err := c.Redis.ZRevRangeByScore(ctx, key, opt)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// ZRevRangeByLex
// https://redis.io/commands/zrevrangebylex
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZRevRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) ([]T, error) {
	mstr, err := c.Redis.ZRevRangeByLex(ctx, key, opt)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// ZScan
// https://redis.io/commands/zscan
// 返回值：member转化为T类型
func (c *ModernRedis[T]) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) (_ []T, nextCursor uint64, _ error) {
	mstr, nextCursor, err := c.Redis.ZScan(ctx, key, cursor, match, count)
	if err != nil {
		return nil, 0, err
	}

	ls, err := c.makeSlice(mstr)
	return ls, nextCursor, err
}

// SPop 随机返回key中的一个member，并从key中删除
// https://redis.io/commands/spop/
// 返回值：转化为T类型
func (c *ModernRedis[T]) SPop(ctx context.Context, key string) (T, error) {
	res, err := c.Redis.SPop(ctx, key, nil)
	if err != nil {
		var nilT T
		return nilT, err
	}

	return c.makeT(res)
}

// SPopN 随机返回key中的count个members，并从key中删除
// https://redis.io/commands/spop/
// 返回值：转化为[]T类型
func (c *ModernRedis[T]) SPopN(ctx context.Context, key string, count int64) ([]T, error) {
	mstr, err := c.Redis.SPopN(ctx, key, count, nil)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// SRandMember 随机返回key中的一个member
// https://redis.io/commands/srandmember/
// 返回值：转化为T类型
func (c *ModernRedis[T]) SRandMember(ctx context.Context, key string) (T, error) {
	res, err := c.Redis.SRandMember(ctx, key, nil)
	if err != nil {
		var nilT T
		return nilT, err
	}
	return c.makeT(res)
}

// SRandMemberN 随机返回key中的count个members
// https://redis.io/commands/srandmember/
// 返回值：转化为[]T类型
func (c *ModernRedis[T]) SRandMemberN(ctx context.Context, key string, count int64) ([]T, error) {
	mstr, err := c.Redis.SRandMemberN(ctx, key, count, nil)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// SScan 扫描set，返回所有匹配的key的members
// https://redis.io/commands/sscan
// 返回值：转化为[]T类型
func (c *ModernRedis[T]) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) (_ []T, nextCursor uint64, _ error) {
	mstr, nextCursor, err := c.Redis.SScan(ctx, key, cursor, match, count)
	if err != nil {
		return nil, 0, err
	}

	ls, err := c.makeSlice(mstr)
	return ls, nextCursor, err
}

// BLPop
// https://redis.io/commands/blpop
// 返回值：key -> T
func (c *ModernRedis[T]) BLPop(ctx context.Context, timeout time.Duration, keys ...string) (map[string]T, error) {
	mstr, err := c.Redis.BLPop(ctx, timeout, nil, keys...)
	if err != nil {
		return nil, err
	}

	return c.makeMap(keys, mstr)
}

// BRPop
// https://redis.io/commands/brpop
// 返回值：key -> T
func (c *ModernRedis[T]) BRPop(ctx context.Context, timeout time.Duration, keys ...string) (map[string]T, error) {
	mstr, err := c.Redis.BRPop(ctx, timeout, nil, keys...)
	if err != nil {
		return nil, err
	}

	return c.makeMap(keys, mstr)
}

// BRPopLPush
// https://redis.io/commands/brpoplpush
// 返回值：转化为T类型
func (c *ModernRedis[T]) BRPopLPush(ctx context.Context, source string, destination string, timeout time.Duration) (T, error) {
	res, err := c.Redis.BRPopLPush(ctx, source, destination, timeout, nil)
	if err != nil {
		var nilT T
		return nilT, err
	}
	return c.makeT(res)
}

// LIndex
// https://redis.io/commands/lindex
// 返回值：转化为T类型
func (c *ModernRedis[T]) LIndex(ctx context.Context, key string, index int64) (T, error) {
	res, err := c.Redis.LIndex(ctx, key, index, nil)
	if err != nil {
		var nilT T
		return nilT, err
	}
	return c.makeT(res)
}

// LPop
// https://redis.io/commands/lpop
// 返回值：转化为T类型
func (c *ModernRedis[T]) LPop(ctx context.Context, key string) (T, error) {
	res, err := c.Redis.LPop(ctx, key, nil)
	if err != nil {
		var nilT T
		return nilT, err
	}
	return c.makeT(res)
}

// LPopCount
// https://redis.io/commands/lpop
// 返回值：转化为[]T类型
func (c *ModernRedis[T]) LPopCount(ctx context.Context, key string, count int) ([]T, error) {
	mstr, err := c.Redis.LPopCount(ctx, key, count, nil)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}

// RPop
// https://redis.io/commands/rpop
// 返回值：转化为T类型
func (c *ModernRedis[T]) RPop(ctx context.Context, key string) (T, error) {
	res, err := c.Redis.RPop(ctx, key, nil)
	if err != nil {
		var nilT T
		return nilT, err
	}
	return c.makeT(res)
}

// RPopCount
// https://redis.io/commands/rpop
// 返回值：转化为[]T类型
func (c *ModernRedis[T]) RPopCount(ctx context.Context, key string, count int) ([]T, error) {
	mstr, err := c.Redis.RPopCount(ctx, key, count, nil)
	if err != nil {
		return nil, err
	}

	return c.makeSlice(mstr)
}
