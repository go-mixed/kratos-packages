package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	// Expiration 缓存过期时间，设置0表示不过期。
	// 注意：设置本option之后仅有如下函数会设置Expiration：
	// 1. Redis.Set/Redis.SetNX/Redis.SetXX/Redis.MSet/Redis.MSetNX
	// 2. Redis.HSet/Redis.HSetNX
	Expiration time.Duration
	// 读写【所有】类型的key都添加一个公共前缀，注意：htable, zset 中的 field/member 和这个无关
	KeyPrefix string
	// SaveEmptyOnRemember remember会保存空值，含nil、空字符串、空数组、空map、空结构体
	SaveEmptyOnRemember bool
	// ForceOnRemember 强制在调用remember时刷新缓存
	ForceOnRemember bool
	// ReturnOriginalKeys 通过 Redis.Scan/Redis.Keys/Redis.ScanType 命令获取到的key是否包含KeyPrefix，即redis中的原始key
	// 为true时，返回的key包含KeyPrefix
	// 为false时，返回的key不包含KeyPrefix
	ReturnOriginalKeys bool
}

func DefaultOptions() Options {
	return Options{
		Expiration:          redis.KeepTTL,
		KeyPrefix:           "",
		SaveEmptyOnRemember: false,
		ForceOnRemember:     false,
		ReturnOriginalKeys:  false,
	}
}

// WithExpiration 设置缓存过期时间，设置0表示不过期。
// 注意：设置本option之后仅有如下函数会设置Expiration：
// 1. Redis.Set/Redis.SetNX/Redis.SetXX/Redis.MSet/Redis.MSetNX
// 2. Redis.HSet/Redis.HSetNX
func (o Options) WithExpiration(expiration time.Duration) Options {
	o.Expiration = expiration
	return o
}

// WithKeyPrefix 读写【所有】类型的key都添加一个公共前缀，注意：htable, zset 中的 field/member 和这个无关
func (o Options) WithKeyPrefix(keyPrefix string) Options {
	o.KeyPrefix = keyPrefix
	return o
}

// WithSaveEmptyOnRemember 调用remember时是否保存空值，含nil、空字符串、空数组、空map、空结构体
func (o Options) WithSaveEmptyOnRemember(saveEmptyOnRemember bool) Options {
	o.SaveEmptyOnRemember = saveEmptyOnRemember
	return o
}

// WithForceOnRemember 调用remember时是否强制刷新缓存
func (o Options) WithForceOnRemember(forceOnRemember bool) Options {
	o.ForceOnRemember = forceOnRemember
	return o
}

// WithReturnOriginalKeys 通过 Redis.Scan/Redis.Keys/Redis.ScanType 命令获取到的key是否包含KeyPrefix，即redis中的原始key
// 为true时，返回的key包含KeyPrefix
// 为false时，返回的key不包含KeyPrefix
func (o Options) WithReturnOriginalKeys(returnOriginalKeys bool) Options {
	o.ReturnOriginalKeys = returnOriginalKeys
	return o
}
