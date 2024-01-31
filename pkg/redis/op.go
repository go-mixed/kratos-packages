package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"time"
)

// Keys 返回所有符合给定 pattern 的 key，pattern支持通配符
// KEYS pattern
// https://redis.io/commands/keys
func (c *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	pattern = c.formatKey(pattern)
	return c.GetRedisCmd(ctx).Keys(ctx, pattern).Result()
}

// Forget 删除缓存，不返回删除的数量
// DEL key [key ...]
// https://redis.io/commands/del
func (c *Redis) Forget(ctx context.Context, keys ...string) error {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).Del(ctx, _keys...).Err()
}

// Del 删除缓存，和 Forget 的区别是，Del返回删除的数量，而Forget不返回
// DEL key [key ...]
// https://redis.io/commands/del
// 返回值：
// 1. int64: 删除的key的数量
// 2. error: 失败时返回的错误
func (c *Redis) Del(ctx context.Context, keys ...string) (int64, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).Del(ctx, _keys...).Result()
}

// Unlink 标记删除，和Del的区别是，Unlink不会阻塞，而Del会阻塞（等内存释放完）
// UNLINK key [key ...]
// https://redis.io/commands/unlink
// 返回值：
// 1. int64: 删除的key的数量
// 2. error: 失败时返回的错误
func (c *Redis) Unlink(ctx context.Context, keys ...string) (int64, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).Unlink(ctx, _keys...).Result()
}

// Dump
// DUMP key
// https://redis.io/commands/dump
func (c *Redis) Dump(ctx context.Context, key string) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Dump(ctx, key).Result()
}

// Migrate 将当前Key迁移到另外一个redis，并指定db、过期时间
// MIGRATE host port key|"" destination-db timeout [COPY] [REPLACE] [KEYS key [key ...]]
// https://redis.io/commands/migrate
func (c *Redis) Migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Migrate(ctx, host, port, key, db, timeout).Result()
}

// Move 将当前db下的key迁移到另一个db
// MOVE key db
// https://redis.io/commands/move
func (c *Redis) Move(ctx context.Context, key string, db int) (bool, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Move(ctx, key, db).Result()
}

func (c *Redis) ObjectRefCount(ctx context.Context, key string) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ObjectRefCount(ctx, key).Result()
}

func (c *Redis) ObjectEncoding(ctx context.Context, key string) *redis.StringCmd {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ObjectEncoding(ctx, key)
}

func (c *Redis) ObjectIdleTime(ctx context.Context, key string) *redis.DurationCmd {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).ObjectIdleTime(ctx, key)
}

// RandomKey 随机返回一个key
// RANDOMKEY
// https://redis.io/commands/randomkey
func (c *Redis) RandomKey(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).RandomKey(ctx).Result()
}

// Rename 将key重命名为newkey
// RENAME key newkey
// https://redis.io/commands/rename
func (c *Redis) Rename(ctx context.Context, key, newkey string) (string, error) {
	key = c.formatKey(key)
	newkey = c.formatKey(newkey)
	return c.GetRedisCmd(ctx).Rename(ctx, key, newkey).Result()
}

// RenameNX 如果newkey不存在，将key重命名为newkey
// RENAME key newkey [NX]
// https://redis.io/commands/renamenx
func (c *Redis) RenameNX(ctx context.Context, key, newkey string) (bool, error) {
	key = c.formatKey(key)
	newkey = c.formatKey(newkey)
	return c.GetRedisCmd(ctx).RenameNX(ctx, key, newkey).Result()
}

func (c *Redis) Restore(ctx context.Context, key string, ttl time.Duration, value string) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Restore(ctx, key, ttl, value).Result()
}

func (c *Redis) RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).RestoreReplace(ctx, key, ttl, value).Result()
}

// Sort 根据排序规则，返回排序后的结果
// SORT key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] [ASC|DESC] [ALPHA] [STORE destination]
// https://redis.io/commands/sort/
//
//	用于 list, set or sorted set
//	actual的结构为&[]Struct{}
func (c *Redis) Sort(ctx context.Context, key string, sort *redis.Sort, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).Sort(ctx, key, sort)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}

	return res.Val(), nil
}

func (c *Redis) SortStore(ctx context.Context, key, store string, sort *redis.Sort) (int64, error) {
	key = c.formatKey(key)
	store = c.formatKey(store)
	return c.GetRedisCmd(ctx).SortStore(ctx, key, store, sort).Result()
}

// SortInterfaces 和Sort的区别是，actual将sort.Get和结果组合成一个struct{}，而Sort的actual是[]Struct{}
// SORT key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] [ASC|DESC] [ALPHA] [STORE destination]
// https://redis.io/commands/sort
//
//	本函数actual的结构非常有局限性，一般用于set、zset，请使用 Sort 来代替
//	actual的结构必须为struct{A int `redis:"a"`, ...}，并且无法内嵌Struct、Map。
func (c *Redis) SortInterfaces(ctx context.Context, key string, sort *redis.Sort, actual any) ([]any, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).SortInterfaces(ctx, key, sort)

	err := utils.IfFunc(actual != nil, func() error { return res.Scan(actual) }, func() error { return res.Err() })
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), nil
}

// Touch 修改key的访问时间
// TOUCH key [key ...]
// https://redis.io/commands/touch
// 返回值：
// 1. int64: 修改的key的数量
// 2. error: 失败时返回的错误
func (c *Redis) Touch(ctx context.Context, keys ...string) (int64, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).Touch(ctx, _keys...).Result()
}

// Type 返回key的类型
// TYPE key
// https://redis.io/commands/type
func (c *Redis) Type(ctx context.Context, key string) (string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).Type(ctx, key).Result()
}

// Copy 将当前key复制到另外db的key，如果replace为true，则会覆盖目标key
// COPY source destination [DB destination-db] [REPLACE]
// https://redis.io/commands/copy
// 参数：
// 1. db: 目标db
// 2. replace: 是否覆盖目标key
// 返回值：
// 1. int64: 1表示复制成功，0表示未复制
// 2. error: 失败时返回的错误
func (c *Redis) Copy(ctx context.Context, sourceKey string, destKey string, db int, replace bool) (int64, error) {
	sourceKey = c.formatKey(sourceKey)
	destKey = c.formatKey(destKey)
	return c.GetRedisCmd(ctx).Copy(ctx, sourceKey, destKey, db, replace).Result()
}

// DebugObject 获取key的调试信息
// DEBUG OBJECT key
// https://redis.io/commands/debug-object
func (c *Redis) DebugObject(ctx context.Context, key string) *redis.StringCmd {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).DebugObject(ctx, key)
}

// MemoryUsage 返回key的内存占用，单位Byte
// MEMORY USAGE key [SAMPLES count]
// https://redis.io/commands/memory-usage
func (c *Redis) MemoryUsage(ctx context.Context, key string, samples ...int) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).MemoryUsage(ctx, key, samples...).Result()
}
