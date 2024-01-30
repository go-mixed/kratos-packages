package redis

import (
	"context"
	"github.com/samber/lo"
)

// SAdd 将members添加到key中
// SADD key member [member ...]
// https://redis.io/commands/sadd/
func (c *Redis) SAdd(ctx context.Context, key string, members ...any) (int64, error) {
	key = c.formatKey(key)
	_members := lo.Map(members, func(v any, _ int) any { return WrapBinaryMarshaler(v) })
	return c.GetRedisCmd(ctx).SAdd(ctx, key, _members...).Result()
}

// SCard 返回key中的member数量，Card->Cardinality
// SCARD key
// https://redis.io/commands/scard/
func (c *Redis) SCard(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SCard(ctx, key).Result()
}

// SDiff 取keys[0]与keys[1:]的差集
// SDIFF key [key ...]
// https://redis.io/commands/sdiff/
func (c *Redis) SDiff(ctx context.Context, keys ...string) ([]string, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).SDiff(ctx, _keys...).Result()
}

// SDiffStore 取keys[0]与keys[1:]的差集，结果存入destination
// SDIFFSTORE destination key [key ...]
// https://redis.io/commands/sdiffstore/
func (c *Redis) SDiffStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	destination = c.formatKey(destination)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).SDiffStore(ctx, destination, _keys...).Result()
}

// SInter 取keys[0]与keys[1:]的交集
// SINTER key [key ...]
// https://redis.io/commands/sinter/
func (c *Redis) SInter(ctx context.Context, keys ...string) ([]string, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).SInter(ctx, _keys...).Result()
}

// SInterStore 取keys[0]与keys[1:]的交集，结果存入destination
// SINTERSTORE destination key [key ...]
// https://redis.io/commands/sinterstore/
func (c *Redis) SInterStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	destination = c.formatKey(destination)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).SInterStore(ctx, destination, _keys...).Result()
}

// SIsMember 判断key中是否存在member
// SISMEMBER key member
// https://redis.io/commands/sismember/
func (c *Redis) SIsMember(ctx context.Context, key string, member any) (bool, error) {
	key = c.formatKey(key) // member是子member，不需要format
	member = WrapBinaryMarshaler(member)
	return c.GetRedisCmd(ctx).SIsMember(ctx, key, member).Result()
}

// SMIsMember 判断key中的members是否存在，返回一个bool的slice，与members一一对应
// SMISMEMBER key member [member ...]
// https://redis.io/commands/sismember/
func (c *Redis) SMIsMember(ctx context.Context, key string, members ...any) ([]bool, error) {
	key = c.formatKey(key) // members是子member，不需要format
	_members := lo.Map(members, func(v any, _ int) any { return WrapBinaryMarshaler(v) })
	return c.GetRedisCmd(ctx).SMIsMember(ctx, key, _members...).Result()
}

// SMembers 以slice形式，返回key中的所有members
// SMEMBERS key
// https://redis.io/commands/smembers/
func (c *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SMembers(ctx, key).Result()
}

// SMembersMap 以map形式，返回key中的所有members
// SMEMBERS key
// https://redis.io/commands/smembers/
func (c *Redis) SMembersMap(ctx context.Context, key string) (map[string]struct{}, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).SMembersMap(ctx, key).Result()
}

// SMove 将source中的member移动到destination中
// SMOVE source destination member
// https://redis.io/commands/smove/
func (c *Redis) SMove(ctx context.Context, source, destination string, member any) (bool, error) {
	source = c.formatKey(source)
	destination = c.formatKey(destination)
	member = WrapBinaryMarshaler(member)
	return c.GetRedisCmd(ctx).SMove(ctx, source, destination, member).Result()
}

// SPop 随机返回key中的一个member，并从key中删除
// SPOP key
// https://redis.io/commands/spop/
func (c *Redis) SPop(ctx context.Context, key string, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).SPop(ctx, key)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), err
}

// SPopN 随机返回key中的count个members，并从key中删除
// SPOP key
// https://redis.io/commands/spop/
func (c *Redis) SPopN(ctx context.Context, key string, count int64, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).SPopN(ctx, key, count)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), err
}

// SRandMember 随机返回key中的一个member
// SRANDMEMBER key
// https://redis.io/commands/srandmember/
func (c *Redis) SRandMember(ctx context.Context, key string, actual any) (string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).SRandMember(ctx, key)

	err := ScanCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return "", filterNil(err)
	}
	return res.Val(), err
}

// SRandMemberN 随机返回key中的count个members
// SRANDMEMBER key [count]
// https://redis.io/commands/srandmember/
func (c *Redis) SRandMemberN(ctx context.Context, key string, count int64, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).SRandMemberN(ctx, key, count)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), err
}

// SRem 删除key中的members。Rem->Remove
// SREM key member [member ...]
// https://redis.io/commands/srem/
// 返回值：
// 1. 被成功移除的元素的数量，不包括未找到的元素。
// 2. error: 失败时返回的错误
func (c *Redis) SRem(ctx context.Context, key string, members ...any) (int64, error) {
	key = c.formatKey(key)
	_members := lo.Map(members, func(v any, _ int) any { return WrapBinaryMarshaler(v) })
	return c.GetRedisCmd(ctx).SRem(ctx, key, _members...).Result()
}

// SUnion 取keys[0]与keys[1:]的并集
// SUNION key [key ...]
// https://redis.io/commands/sunion/
func (c *Redis) SUnion(ctx context.Context, keys ...string) ([]string, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).SUnion(ctx, _keys...).Result()
}

// SUnionStore 取keys[0]与keys[1:]的并集，结果存入destination
// SUNIONSTORE destination key [key ...]
// https://redis.io/commands/sunionstore/
func (c *Redis) SUnionStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	destination = c.formatKey(destination)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).SUnionStore(ctx, destination, _keys...).Result()
}
