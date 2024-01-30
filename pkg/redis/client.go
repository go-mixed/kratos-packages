package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// Context returns the context of the client
func (c *Redis) Context() context.Context {
	return c.originalClient.Context()
}

// WithContext returns a shallow copy of c with its context changed
func (c *Redis) WithContext(ctx context.Context) *Redis {
	_c := c.Clone()
	_c.originalClient = c.originalClient.WithContext(ctx)
	return _c
}

// AddHook adds a hook to the client.
func (c *Redis) AddHook(hook redis.Hook) {
	c.originalClient.AddHook(hook)
}

// Watch watches the given keys for modifications and executes the fn
func (c *Redis) Watch(ctx context.Context, fn func(*Tx) error, keys ...string) error {
	_keys := c.formatKeys(keys)
	return c.originalClient.Watch(ctx, func(tx *redis.Tx) error {
		return fn(NewTx(c, tx, c.options))
	}, _keys...)
}

// Do executes a command
func (c *Redis) Do(ctx context.Context, args ...any) *redis.Cmd {
	return c.originalClient.Do(ctx, args...)
}

// Process processes a command
func (c *Redis) Process(ctx context.Context, cmd redis.Cmder) error {
	return c.originalClient.Process(ctx, cmd)
}

// Close closes the client
func (c *Redis) Close() error {
	return c.originalClient.Close()
}

// PoolStats returns the pool stats
func (c *Redis) PoolStats() *redis.PoolStats {
	return c.originalClient.PoolStats()
}
