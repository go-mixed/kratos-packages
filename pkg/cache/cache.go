package cache

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"time"
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

// ServerTimeDelta 获取redis服务器时间与本地时间的差值（注意：会有socket传输时间的误差）
// e.g.: time.Now().Add(ServerTimeDelta()) 可以得到redis服务器的时间
func (c *Cache) ServerTimeDelta(ctx context.Context) time.Duration {
	serverTime, err := c.predis.Time(ctx)
	if err != nil || serverTime.IsZero() {
		return 0
	}

	return serverTime.Sub(time.Now())
}

func (c *Cache) WithOptions(options redis.Options) *Cache {
	return &Cache{
		predis:  c.predis.WithOptions(options),
		logger:  c.logger,
		options: options,
	}
}

// WithKeyPrefix 设置key前缀，并返回新的Cache
func (c *Cache) WithKeyPrefix(keyPrefix string) *Cache {
	options := c.options
	options.KeyPrefix = keyPrefix
	return c.WithOptions(options)
}

// WithExpiration 设置过期时间，并返回新的Cache，0表示不过期
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

// Default 并返回使用默认值新的Cache。
// 默认值为： keyPrefix: "", expiration: -1, saveEmptyOnRemember: false
func (c *Cache) Default() *Cache {
	options := redis.DefaultOptions()
	return c.WithOptions(options)
}
