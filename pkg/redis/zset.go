package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"time"
)

// BZPopMax 阻塞zset，从zset中取出score最大的元素或者等待超时，会删除元素
// BZPOPMAX key [key ...] timeout
// https://redis.io/commands/bzpopmax
// 参数：
// 1. timeout: 超时时间
// 返回值：
// 1. *redis.ZWithKey: zset的key、member、score
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) (*redis.ZWithKey, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).BZPopMax(ctx, timeout, _keys...).Result()
}

// BZPopMin 阻塞zset，从zset中取出score最小的元素或者等待超时，会删除元素
// BZPOPMIN key [key ...] timeout
// https://redis.io/commands/bzpopmin
// 参数：
// 1. timeout: 超时时间
// 返回值：
// 1. *redis.ZWithKey: zset的key、member、score
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) (*redis.ZWithKey, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).BZPopMin(ctx, timeout, _keys...).Result()
}

func (c *Redis) formatPZ(members []*redis.Z) []*redis.Z {
	return lo.Map(members, func(v *redis.Z, _ int) *redis.Z {
		if v == nil {
			return nil
		}
		return &redis.Z{Member: WrapBinaryMarshaler(v.Member), Score: v.Score}
	})
}
func (c *Redis) formatZ(members []redis.Z) []redis.Z {
	return lo.Map(members, func(v redis.Z, _ int) redis.Z {
		return redis.Z{Member: WrapBinaryMarshaler(v.Member), Score: v.Score}
	})
}

// ZAdd 将一个或多个 member 元素及其 score 值加入到有序集 key 当中。
// ZADD key score member [score member ...]
// https://redis.io/commands/zadd
func (c *Redis) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	key = c.formatKey(key)
	_members := c.formatPZ(members)
	return c.GetRedisCmd(ctx).ZAdd(ctx, key, _members...).Result()
}

// ZAddNX 将一个或多个 member 元素及其 score 值加入到有序集 key 当中，只有当member不存在时才会添加
// ZADD key NX score member [score member ...]
// https://redis.io/commands/zadd
// 返回值：
// 1. int64: 被成功添加的新成员的数量
// 2. error: 失败时返回的错误
func (c *Redis) ZAddNX(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	key = c.formatKey(key)
	_members := c.formatPZ(members)
	return c.GetRedisCmd(ctx).ZAddNX(ctx, key, _members...).Result()
}

// ZAddXX 将一个或多个 member 元素及其 score 值加入到有序集 key 当中，只有当member存在时才会添加
// ZADD key XX score member [score member ...]
// https://redis.io/commands/zadd
// 返回值：
// 1. int64: 为0（因为没有新添加的成员）
// 2. error: 失败时返回的错误
func (c *Redis) ZAddXX(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	key = c.formatKey(key)
	_members := c.formatPZ(members)
	return c.GetRedisCmd(ctx).ZAddXX(ctx, key, _members...).Result()
}

// ZAddCh 将一个或多个 member 元素及其 score 值加入到有序集 key 当中
// ZADD key CH score member [score member ...]
// https://redis.io/commands/zadd
// 返回值：
// 1. int64: 被成功添加的、更新的成员的数量
// 2. error: 失败时返回的错误
func (c *Redis) ZAddCh(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	key = c.formatKey(key)
	_members := c.formatPZ(members)
	return c.GetRedisCmd(ctx).ZAddCh(ctx, key, _members...).Result()
}

// ZAddNXCh 将一个或多个 member 元素及其 score 值加入到有序集 key 当中，只有当member不存在时才会添加
// ZADD key NX CH score member [score member ...]
// https://redis.io/commands/zadd
// 返回值：
// 1. int64: 被成功添加的新成员的数量
// 2. error: 失败时返回的错误
func (c *Redis) ZAddNXCh(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	key = c.formatKey(key)
	_members := c.formatPZ(members)
	return c.GetRedisCmd(ctx).ZAddNXCh(ctx, key, _members...).Result()
}

// ZAddXXCh 将一个或多个 member 元素及其 score 值加入到有序集 key 当中，只有当member存在时才会添加
// ZADD key XX CH score member [score member ...]
// https://redis.io/commands/zadd
// 返回值：
// 1. int64: score被修改的数量
// 2. error: 失败时返回的错误
func (c *Redis) ZAddXXCh(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	key = c.formatKey(key)
	_members := c.formatPZ(members)
	return c.GetRedisCmd(ctx).ZAddXXCh(ctx, key, _members...).Result()
}

// ZAddArgs 传递ZAddArgs参数的方式添加元素
// ZADD key [NX|XX] [CH] score member [score member ...]
// https://redis.io/commands/zadd
// 返回值：
// 1. int64: 被成功添加的新成员的数量；如果CH为true，则返回被成功添加或者被更新的成员数量
func (c *Redis) ZAddArgs(ctx context.Context, key string, args redis.ZAddArgs) (int64, error) {
	key = c.formatKey(key)
	args.Members = c.formatZ(args.Members)
	return c.GetRedisCmd(ctx).ZAddArgs(ctx, key, args).Result()
}

// ZAddArgsIncr 传递ZAddArgs参数的方式添加元素，score+1。args.Members中的score不会被使用
// ZADD key [NX|XX] [CH] [INCR] score member [score member ...]
// https://redis.io/commands/zadd
// 返回值：
// 1. float64: member成员+1后的score
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) ZAddArgsIncr(ctx context.Context, key string, args redis.ZAddArgs) (float64, error) {
	key = c.formatKey(key)
	args.Members = c.formatZ(args.Members)
	return c.GetRedisCmd(ctx).ZAddArgsIncr(ctx, key, args).Result()
}

// ZIncr 将member成员的score值加1
// ZINCRBY key increment member
// https://redis.io/commands/zincrby
// 返回值：
// 1. float64: member成员+1后的score
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) ZIncr(ctx context.Context, key string, member *redis.Z) (float64, error) {
	key = c.formatKey(key) // member是子member，不需要format
	if member == nil {
		return c.GetRedisCmd(ctx).ZIncr(ctx, key, nil).Result()
	}
	_member := *member
	_member.Member = WrapBinaryMarshaler(member.Member)
	return c.GetRedisCmd(ctx).ZIncr(ctx, key, &_member).Result()
}

// ZIncrNX 将member成员的score值加1，只有当member不存在时才会+1。member.Score不会被使用
// ZAdd key NX INCR member
// https://redis.io/commands/zincrby
// 返回值：
// 1. float64: member成员+1后的score，为1表示member不存在，为0
func (c *Redis) ZIncrNX(ctx context.Context, key string, member *redis.Z) (float64, error) {
	key = c.formatKey(key) // member是子member，不需要format
	if member == nil {
		return c.GetRedisCmd(ctx).ZIncrNX(ctx, key, nil).Result()
	}
	_member := *member
	_member.Member = WrapBinaryMarshaler(member.Member)
	return c.GetRedisCmd(ctx).ZIncrNX(ctx, key, &_member).Result()
}

// ZIncrXX 将member成员的score值加1，只有当member存在时才会+1。member.Score不会被使用
// ZAdd key XX INCR member
// https://redis.io/commands/zincrby
func (c *Redis) ZIncrXX(ctx context.Context, key string, member *redis.Z) (float64, error) {
	key = c.formatKey(key) // member是子member，不需要format
	if member == nil {
		return c.GetRedisCmd(ctx).ZIncrNX(ctx, key, nil).Result()
	}
	_member := *member
	_member.Member = WrapBinaryMarshaler(member.Member)
	return c.GetRedisCmd(ctx).ZIncrXX(ctx, key, &_member).Result()
}

// ZCard 返回有序集 key 的基数。Card->Cardinality
// ZCARD key
// https://redis.io/commands/zcard
func (c *Redis) ZCard(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZCard(ctx, key).Result()
}

// ZCount 返回有序集 key 中，score 值在 min 和 max 之间(均包含)的成员的数量
// ZCOUNT key min max
// https://redis.io/commands/zcount
// 参数：
// 1. min: 最小score，-inf表示无穷小
// 2. max: 最大score，+inf表示无穷大
func (c *Redis) ZCount(ctx context.Context, key, min, max string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZCount(ctx, key, min, max).Result()
}

// ZLexCount 返回有序集 key 中，成员的 score 值在 min 和 max 之间(包含)，并且member按照字典顺序排列的成员的数量
// ZLEXCOUNT key min max
// https://redis.io/commands/zlexcount
func (c *Redis) ZLexCount(ctx context.Context, key, min, max string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZLexCount(ctx, key, min, max).Result()
}

// ZIncrBy 将member成员的score值加上increment
// ZINCRBY key increment member
// https://redis.io/commands/zincrby
// 返回值：
// 1. float64: member成员+increment后的score
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZIncrBy(ctx, key, increment, member).Result()
}

// ZInter 指定keys和weights，计算交集，如果有传递aggregate，则会计算聚合值
// ZINTER numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
func (c *Redis) ZInter(ctx context.Context, store *redis.ZStore) ([]string, error) {
	if store == nil {
		return c.GetRedisCmd(ctx).ZInter(ctx, nil).Result()
	}
	_store := *store
	_store.Keys = c.formatKeys(store.Keys)
	return c.GetRedisCmd(ctx).ZInter(ctx, &_store).Result()
}

// ZInterWithScores 指定keys和weights，计算交集，如果有传递aggregate，则会计算聚合值。返回值包含member、score
// ZINTER numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX] [WITHSCORES]
// https://redis.io/commands/zinter
func (c *Redis) ZInterWithScores(ctx context.Context, store *redis.ZStore) ([]redis.Z, error) {
	if store == nil {
		return c.GetRedisCmd(ctx).ZInterWithScores(ctx, nil).Result()
	}
	_store := *store
	_store.Keys = c.formatKeys(store.Keys)
	return c.GetRedisCmd(ctx).ZInterWithScores(ctx, &_store).Result()
}

// ZInterStore 指定keys和weights，计算交集，如果有传递aggregate，则会计算聚合值。将结果存入destination
// ZINTERSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
// https://redis.io/commands/zinterstore
// 返回值：
// 1. int64: 交集的数量
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) ZInterStore(ctx context.Context, destination string, store *redis.ZStore) (int64, error) {
	destination = c.formatKey(destination)
	if store == nil {
		return c.GetRedisCmd(ctx).ZInterStore(ctx, destination, nil).Result()
	}
	_store := *store
	_store.Keys = c.formatKeys(store.Keys)
	return c.GetRedisCmd(ctx).ZInterStore(ctx, destination, &_store).Result()
}

// ZMScore 返回有序集 key 中，一个或多个成员的 score 值
// ZMSCORE key member [member ...]
// https://redis.io/commands/zmscore
func (c *Redis) ZMScore(ctx context.Context, key string, members ...string) ([]float64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZMScore(ctx, key, members...).Result()
}

// ZPopMax 移除并返回有序集合中按score倒序排序TOP N元素。不传count表示1个元素，传递多个count以第一个为准
// ZPOPMAX key [count]
// https://redis.io/commands/zpopmax
func (c *Redis) ZPopMax(ctx context.Context, key string, count ...int64) ([]redis.Z, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZPopMax(ctx, key, count...).Result()
}

// ZPopMin 移除并返回有序集合中按score正序排序TOP N元素。不传count表示1个元素，传递多个count以第一个为准
// ZPOPMIN key [count]
// https://redis.io/commands/zpopmin
func (c *Redis) ZPopMin(ctx context.Context, key string, count ...int64) ([]redis.Z, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZPopMin(ctx, key, count...).Result()
}

// ZRange 返回有序集 key 中，score在start-stop（均包含）的member列表
// ZRANGE key start stop
// https://redis.io/commands/zrange
func (c *Redis) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 返回有序集 key 中，score在start-stop（均包含）的member、score列表
// ZRANGE key start stop [WITHSCORES]
// https://redis.io/commands/zrange
func (c *Redis) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRangeByScore 传递Min/Max/Offset/Count参数的方式返回有序集 key 中，score在min-max（均包含）的member列表
// ZRANGEBYSCORE key min max [LIMIT offset count]
// https://redis.io/commands/zrangebyscore
// min、max可以设置为-inf、+inf
func (c *Redis) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRangeByScore(ctx, key, opt).Result()
}

// ZRangeByLex 传递Min/Max/Offset/Count参数的方式返回有序集 key 中，score在min-max（均包含）的member列表，member按照字典顺序排列
// ZRANGEBYLEX key min max [LIMIT offset count]
// https://redis.io/commands/zrangebylex
// min、max可以设置为-inf、+inf
func (c *Redis) ZRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRangeByLex(ctx, key, opt).Result()
}

// ZRangeByScoreWithScores 传递Min/Max/Offset/Count参数的方式返回有序集 key 中，score在min-max（均包含）的member、score列表
// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
// https://redis.io/commands/zrangebyscore
// min、max可以设置为-inf、+inf
func (c *Redis) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRangeByScoreWithScores(ctx, key, opt).Result()
}

// ZRangeArgs 传递ZRangeArgs参数的方式返回有序集 key 中，score在min-max（均包含）的member列表
// ZRANGE key start stop [BYSCORE | BYLEX] [REV] [LIMIT offset count]
// https://redis.io/commands/zrange
// min、max可以设置为-inf、+inf
func (c *Redis) ZRangeArgs(ctx context.Context, z redis.ZRangeArgs) ([]string, error) {
	_z := z
	_z.Key = c.formatKey(z.Key)
	return c.GetRedisCmd(ctx).ZRangeArgs(ctx, _z).Result()
}

// ZRangeArgsWithScores 传递ZRangeArgs参数的方式返回有序集 key 中，score在min-max（均包含）的member、score列表
// ZRANGE key start stop [BYSCORE | BYLEX] [REV] [LIMIT offset count] [WITHSCORES]
// https://redis.io/commands/zrange
func (c *Redis) ZRangeArgsWithScores(ctx context.Context, z redis.ZRangeArgs) ([]redis.Z, error) {
	_z := z
	_z.Key = c.formatKey(z.Key)
	return c.GetRedisCmd(ctx).ZRangeArgsWithScores(ctx, _z).Result()
}

// ZRangeStore 传递ZRangeArgs参数的方式返回有序集 key 中，score在min-max（均包含）的member列表，并将结果存入destination
// ZRANGESTORE dst src min max [BYSCORE | BYLEX] [REV] [LIMIT offset count]
// https://redis.io/commands/zrangestore
func (c *Redis) ZRangeStore(ctx context.Context, dst string, z redis.ZRangeArgs) (int64, error) {
	dst = c.formatKey(dst)
	_z := z
	_z.Key = c.formatKey(z.Key)
	return c.GetRedisCmd(ctx).ZRangeStore(ctx, dst, _z).Result()
}

// ZRank 返回有序集 key 中成员 member 的排名（从0开始）。其中有序集成员按 score 值递增(从小到大)顺序排列
// ZRANK key member
// https://redis.io/commands/zrank
func (c *Redis) ZRank(ctx context.Context, key, member string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRank(ctx, key, member).Result()
}

// ZRem 移除有序集 key 中的一个或多个成员，不存在的成员将被忽略。Rem->Remove
// ZREM key member [member ...]
// https://redis.io/commands/zrem
func (c *Redis) ZRem(ctx context.Context, key string, members ...any) (int64, error) {
	key = c.formatKey(key)
	_members := lo.Map(members, func(v any, _ int) any { return WrapBinaryMarshaler(v) })
	return c.GetRedisCmd(ctx).ZRem(ctx, key, _members...).Result()
}

// ZRemRangeByRank 移除有序集 key 中，指定排名(rank)区间内的所有成员。
// ZREMRANGEBYRANK key start stop
// https://redis.io/commands/zremrangebyrank
func (c *Redis) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRemRangeByRank(ctx, key, start, stop).Result()
}

// ZRemRangeByScore 移除有序集 key 中，指定score区间内的所有成员。
// ZREMRANGEBYSCORE key min max
// https://redis.io/commands/zremrangebyscore
// min、max可以设置为-inf、+inf
func (c *Redis) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRemRangeByScore(ctx, key, min, max).Result()
}

// ZRemRangeByLex 移除有序集 key 中，指定member区间内的所有成员，member按照字典顺序排列。
func (c *Redis) ZRemRangeByLex(ctx context.Context, key, min, max string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRemRangeByLex(ctx, key, min, max).Result()
}

// ZRevRange 返回有序集 key 中，score在start-stop（均包含）的member列表，按照score倒序排列
// ZREVRANGE key start stop
// https://redis.io/commands/zrevrange
func (c *Redis) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRevRange(ctx, key, start, stop).Result()
}

// ZRevRangeWithScores 返回有序集 key 中，score在start-stop（均包含）的member、score列表，按照score倒序排列
// ZREVRANGE key start stop [WITHSCORES]
// https://redis.io/commands/zrevrange
func (c *Redis) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRevRangeWithScores(ctx, key, start, stop).Result()
}

// ZRevRangeByScore 传递Min/Max/Offset/Count参数的方式返回有序集 key 中，score在min-max（均包含）的member列表，按照score倒序排列
// ZREVRANGEBYSCORE key max min [LIMIT offset count]
// https://redis.io/commands/zrevrangebyscore
// min、max可以设置为-inf、+inf
func (c *Redis) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRevRangeByScore(ctx, key, opt).Result()
}

// ZRevRangeByLex 传递Min/Max/Offset/Count参数的方式返回有序集 key 中，score在min-max（均包含）的member列表，按照score倒序排列。member按照字典顺序排列
// ZREVRANGEBYLEX key max min [LIMIT offset count]
// https://redis.io/commands/zrevrangebylex
// min、max可以设置为-inf、+inf
func (c *Redis) ZRevRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRevRangeByLex(ctx, key, opt).Result()
}

// ZRevRangeByScoreWithScores 传递Min/Max/Offset/Count参数的方式返回有序集 key 中，score在min-max（均包含）的member、score列表，按照score倒序排列
// ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count]
// https://redis.io/commands/zrevrangebyscore
// min、max可以设置为-inf、+inf
func (c *Redis) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRevRangeByScoreWithScores(ctx, key, opt).Result()
}

// ZRevRank 返回有序集 key 中成员 member 的排名。其中有序集成员按 score 值递减(从大到小)顺序排列
// ZREVRANK key member
// https://redis.io/commands/zrevrank
func (c *Redis) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRevRank(ctx, key, member).Result()
}

// ZScore 返回有序集 key 中，成员 member 的 score 值
// ZSCORE key member
// https://redis.io/commands/zscore
func (c *Redis) ZScore(ctx context.Context, key, member string) (float64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZScore(ctx, key, member).Result()
}

// ZUnionStore 指定keys和weights，计算并集，如果有传递aggregate，则会计算聚合值。将结果存入destination
// ZUNIONSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
// https://redis.io/commands/zunionstore
func (c *Redis) ZUnionStore(ctx context.Context, dest string, store *redis.ZStore) (int64, error) {
	dest = c.formatKey(dest)
	if store == nil {
		return c.GetRedisCmd(ctx).ZUnionStore(ctx, dest, nil).Result()
	}
	_store := *store
	_store.Keys = c.formatKeys(store.Keys)
	return c.GetRedisCmd(ctx).ZUnionStore(ctx, dest, &_store).Result()
}

// ZUnion 指定keys和weights，计算并集，如果有传递aggregate，则会计算聚合值
// ZUNION numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
// https://redis.io/commands/zunion
func (c *Redis) ZUnion(ctx context.Context, store redis.ZStore) ([]string, error) {
	_store := store
	_store.Keys = c.formatKeys(store.Keys)
	return c.GetRedisCmd(ctx).ZUnion(ctx, _store).Result()
}

// ZUnionWithScores 指定keys和weights，计算并集，如果有传递aggregate，则会计算聚合值。返回值包含member、score
// ZUNION numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX] [WITHSCORES]
// https://redis.io/commands/zunion
func (c *Redis) ZUnionWithScores(ctx context.Context, store redis.ZStore) ([]redis.Z, error) {
	_store := store
	_store.Keys = c.formatKeys(store.Keys)
	return c.GetRedisCmd(ctx).ZUnionWithScores(ctx, _store).Result()
}

// ZRandMember 返回有序集 key 中，随机获取count个元素。
// ZRANDMEMBER key [count]
// https://redis.io/commands/zrandmember
// 参数：
// 1. count：如果count为负数数，则返回的元素中可能包含重复元素；如果count为正数，则返回的元素中不包含重复元素
// 返回值：
// 1. []string: 随机获取的元素列表。如果withScores为true，则返回的[]string{member1, score1, member2, score2, ...}，否则返回[]string{member1, member2, ...}
// 2. error: 失败时返回的错误，不会返回redis.Nil
func (c *Redis) ZRandMember(ctx context.Context, key string, count int, withScores bool) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ZRandMember(ctx, key, count, withScores).Result()
}

// ZDiff 返回有序集 key1 中，存在于 key1 且不存在于 key2、key3... 的member列表
// ZDIFF key [key ...]
// https://redis.io/commands/zdiff
func (c *Redis) ZDiff(ctx context.Context, keys ...string) ([]string, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).ZDiff(ctx, _keys...).Result()
}

// ZDiffWithScores 返回有序集 key1 中，存在于 key1 且不存在于 key2、key3... 的member、score列表
// ZDIFF key [key ...] [WITHSCORES]
// https://redis.io/commands/zdiff
func (c *Redis) ZDiffWithScores(ctx context.Context, keys ...string) ([]redis.Z, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).ZDiffWithScores(ctx, _keys...).Result()
}

// ZDiffStore 将有序集 key1 中，存在于 key1 且不存在于 key2、key3... 的member、score列表存入destination
// ZDIFFSTORE destination key [key ...]
// https://redis.io/commands/zdiffstore
func (c *Redis) ZDiffStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	destination = c.formatKey(destination)
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).ZDiffStore(ctx, destination, _keys...).Result()
}
