package cache

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"time"
)

type modernCache[T any] struct {
	*redis.ModernRedis[T]
	logger  log.Logger
	options redis.Options
}

// AsModernCache 直接定义T，并返回ModernRedis[T]。比如：AsModernCache[string](c)，这是显式的定义T
func AsModernCache[T any](c *Cache) *modernCache[T] {
	return &modernCache[T]{
		ModernRedis: redis.NewModernRedis[T](c.predis),
		logger:      c.logger,
		options:     c.options,
	}
}

// AsModernCacheBy 使用actual来定义T，并返回ModernRedis[T]。
// 和AsModernCache得到的结果是一样的，只是用参数来推导T，比如：AsModernCacheBy(c, "")，T会被推导为string
func AsModernCacheBy[T any](c *Cache, actual T) *modernCache[T] {
	return AsModernCache[T](c)
}

func (c *modernCache[T]) hook() {
	c.ModernRedis.AddHook(newCacheHook(c.options, c.logger))
}

// Clone 克隆一个modernCache，ModernRedis也会被克隆
// 注意：所有hook都会被克隆
func (c *modernCache[T]) Clone() *modernCache[T] {
	return &modernCache[T]{
		ModernRedis: redis.NewModernRedis[T](c.ModernRedis.Redis.Clone()),
		logger:      c.logger,
		options:     c.options,
	}
}

func (c *modernCache[T]) WithOptions(options redis.Options) *modernCache[T] {
	return &modernCache[T]{
		ModernRedis: redis.NewModernRedis[T](c.ModernRedis.Redis.WithOptions(options)),
		logger:      c.logger,
		options:     options,
	}
}

// WithKeyPrefix 设置key前缀，并返回新的Cache
func (c *modernCache[T]) WithKeyPrefix(keyPrefix string) *modernCache[T] {
	options := c.options
	options.KeyPrefix = keyPrefix

	return c.WithOptions(options)
}

// WithExpiration 设置过期时间，并返回新的Cache，0表示不过期
func (c *modernCache[T]) WithExpiration(expiration time.Duration) *modernCache[T] {
	options := c.options
	options.Expiration = expiration

	return c.WithOptions(options)
}

// WithSaveEmptyOnRemember 设置在调用remember时是否保存空值（nil、空字符串、空数组、空map、空结构体），并返回新的Cache
func (c *modernCache[T]) WithSaveEmptyOnRemember(saveIfZero bool) *modernCache[T] {
	options := c.options
	options.SaveEmptyOnRemember = saveIfZero

	return c.WithOptions(options)
}

// WithForceOnRemember 设置在调用remember时是否强制刷新缓存，并返回新的Cache
func (c *modernCache[T]) WithForceOnRemember(force bool) *modernCache[T] {
	options := c.options
	options.ForceOnRemember = force

	return c.WithOptions(options)
}

// WithRedis 设置redis客户端，并返回新的Cache
func (c *modernCache[T]) WithRedis(client *redis.Client) *modernCache[T] {
	_c := &modernCache[T]{
		ModernRedis: redis.NewModernRedis[T](redis.NewRedis(client, c.options)),
		logger:      c.logger,
		options:     c.options,
	}
	_c.hook() // 新的client，需要重新注册hook
	return _c
}

// Default 并返回使用默认值新的Cache。
// 默认值为： keyPrefix: "", expiration: -1, saveEmptyOnRemember: false
func (c *modernCache[T]) Default() *modernCache[T] {
	options := redis.DefaultOptions()
	return c.WithOptions(options)
}
