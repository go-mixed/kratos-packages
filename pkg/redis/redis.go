package redis

import (
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.uber.org/multierr"
)

type Redis struct {
	originalClient *Client

	options Options
}

func NewRedis(client *redis.Client, options Options) *Redis {
	c := &Redis{
		originalClient: client,
		options:        options,
	}
	return c
}

// Clone 克隆一个Redis，originalClient也会被克隆
func (c *Redis) Clone() *Redis {
	return &Redis{
		originalClient: c.originalClient.WithTimeout(c.originalClient.Options().ReadTimeout),
		options:        c.options,
	}
}

// WithOptions 重新设置options，并返回新的Redis
func (c *Redis) WithOptions(options Options) *Redis {
	_c := c.Clone()
	_c.options = options
	return _c
}

// formatKey 格式化key，加上前缀
func (c *Redis) formatKey(key string) string {
	return c.options.KeyPrefix + key
}

// formatKeys 格式化keys，加上前缀
func (c *Redis) formatKeys(keys []string) []string {
	return lo.Map(keys, func(key string, _ int) string {
		return c.formatKey(key)
	})
}

// formatMKeys 格式化map的key，加上前缀，value会被WrapBinaryMarshaler
func (c *Redis) formatMKeys(kvs map[string]any) map[string]any {
	return lo.MapEntries(kvs, func(k string, v any) (string, any) {
		return c.formatKey(k), WrapBinaryMarshaler(v)
	})
}

// expire 设置过期时间，对于所有设置类的操作都会调用该方法
func (c *Redis) expire(ctx context.Context, err error, keys ...string) error {
	if err != nil {
		return err
	}

	// KeepTTL 表示不修改过期时间
	if c.options.Expiration == redis.KeepTTL {
		return nil
	}

	for _, key := range keys {
		err = multierr.Append(err,
			c.GetRedisCmd(ctx).Expire(ctx, key, c.options.Expiration).Err(),
		)
	}
	return err
}

// GetRedisCmd 获取redis的命令执行器，如果在上下文中有pipeliner，则返回pipeliner，否则返回redis.Client
func (c *Redis) GetRedisCmd(ctx context.Context) redis.Cmdable {
	if pipeliner, ok := fromContext(ctx); ok {
		return pipeliner
	}

	return c.originalClient
}

// filterNil 过滤掉redis.Nil的错误
func filterNil(err error) error {
	if err != nil && !errors.Is(err, Nil) {
		return err
	}
	return nil
}
