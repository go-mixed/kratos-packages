package cache

import (
	"time"

	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
)

type predis = redis.Redis

type Cache struct {
	*predis // 防止外部修改Cache.Redis，但是又可以使用Cache.Gedis的方法
	logger  log.Logger
	options redis.Options
}

func NewCache(
	client *redis.Client,
	logger log.Logger,
	options ...Option) *Cache {
	c := &Cache{
		logger: logger,
	}

	for _, option := range options {
		option(c)
	}

	// 赋值的是client的副本，hook时不会修改到外部的client
	c.predis = redis.NewRedis(client.WithTimeout(client.Options().ReadTimeout), c.options)
	c.hook()
	return c
}

func (c *Cache) hook() {
	c.predis.AddHook(newCacheHook(c.options, c.logger))
}

// Clone 克隆一个Cache，predis也会被克隆
// 注意：所有hook都会被克隆
func (c *Cache) Clone() *Cache {
	return &Cache{
		predis:  c.predis.Clone(),
		options: c.options,
		logger:  c.logger,
	}
}

func (c *Cache) GetOptions() redis.Options {
	return c.options
}

func (c *Cache) WithOptions(options redis.Options) *Cache {
	return &Cache{
		predis:  c.predis.WithOptions(options),
		logger:  c.logger,
		options: options,
	}
}

// WithKeyPrefix 设置key前缀，并返回新的Cache
// 给【所有】类型的key添加公共前缀，注意：htable, zset 中的 field/member 和这个无关
func (c *Cache) WithKeyPrefix(keyPrefix string) *Cache {
	options := c.options
	options.KeyPrefix = keyPrefix
	return c.WithOptions(options)
}

// WithExpiration 设置过期时间，并返回新的Cache，0表示不过期
// 注意：设置本option之后仅有如下函数会设置Expiration：
// 1. Set/SetNX/SetXX/SetEX/SetPX
// 2. HSet/HSetNX/HSetXX/HSetEX/HSetPX
func (c *Cache) WithExpiration(expiration time.Duration) *Cache {
	options := c.options
	options.Expiration = expiration
	return c.WithOptions(options)
}

// WithSaveEmptyOnRemember 设置在调用remember时是否保存空值，并返回新的Cache
func (c *Cache) WithSaveEmptyOnRemember(saveIfZero bool) *Cache {
	options := c.options
	options.SaveEmptyOnRemember = saveIfZero
	return c.WithOptions(options)
}

// WithReturnOriginalKeys 通过 Scan/Keys/ScanType 命令获取到的key是否包含KeyPrefix，即redis中的原始key
// 为true时，返回的key包含KeyPrefix
// 为false时，返回的key不包含KeyPrefix
func (c *Cache) WithReturnOriginalKeys(returnOriginalKeys bool) *Cache {
	options := c.options
	options.ReturnOriginalKeys = returnOriginalKeys
	return c.WithOptions(options)
}

// WithRedis 设置redis客户端，并返回新的Cache
func (c *Cache) WithRedis(client *redis.Client) *Cache {
	_c := &Cache{
		predis:  redis.NewRedis(client, c.options),
		logger:  c.logger,
		options: c.options,
	}
	_c.hook() // 新的client，需要重新注册hook
	return _c
}

// GetRedis 获取Redis
func (c *Cache) GetRedis() *redis.Redis {
	return c.predis
}

// Default 并返回使用默认值新的Cache。
// 默认值为： keyPrefix: "", expiration: -1, saveEmptyOnRemember: false
func (c *Cache) Default() *Cache {
	options := redis.DefaultOptions()
	return c.WithOptions(options)
}
