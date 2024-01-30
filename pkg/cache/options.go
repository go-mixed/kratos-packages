package cache

import (
	"time"
)

type Option func(*Cache)

// WithExpiration 设置过期时间
func WithExpiration(expiration time.Duration) func(*Cache) {
	return func(c *Cache) {
		c.options.Expiration = expiration
	}
}

// WithKeyPrefix 设置key前缀
func WithKeyPrefix(keyPrefix string) func(*Cache) {
	return func(c *Cache) {
		c.options.KeyPrefix = keyPrefix
	}
}

// WithSaveEmptyOnRemember 设置在调用remember时是否保存空值
func WithSaveEmptyOnRemember(saveIfZero bool) func(*Cache) {
	return func(c *Cache) {
		c.options.SaveEmptyOnRemember = saveIfZero
	}
}
