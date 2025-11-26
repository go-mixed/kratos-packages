package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Expiration time.Duration
	KeyPrefix  string
	// SaveEmptyOnRemember remember会保存空值，含nil、空字符串、空数组、空map、空结构体
	SaveEmptyOnRemember bool
	// ForceOnRemember 强制在remember时刷新缓存
	ForceOnRemember bool
	// ReturnOriginalKeys 通过 Redis.Scan Redis.Keys Redis.ScanType 命令获取到的key是否包含前缀，为true时，返回的key包含前缀，为false时，返回的key不包含前缀
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

func (o Options) WithExpiration(expiration time.Duration) Options {
	o.Expiration = expiration
	return o
}

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

// WithReturnOriginalKeys 通过 Redis.Scan Redis.Keys Redis.ScanType 命令获取到的key是否包含前缀，为true时，返回的key包含前缀，为false时，返回的key不包含前缀
func (o Options) WithReturnOriginalKeys(returnOriginalKeys bool) Options {
	o.ReturnOriginalKeys = returnOriginalKeys
	return o
}
