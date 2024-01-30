package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Broadcaster interface {
	Publish(ctx context.Context, channel string, message any) (int64, error)
	SubscribeHandler(ctx context.Context, handler func(ctx context.Context, message string), channels ...string) error
}

var _ Broadcaster = (*Redis)(nil)

// Publish 发布一个消息到指定的频道channel
// PUBLISH channel message
// https://redis.io/commands/publish
// 返回值：
// 1. int64: 接收到消息的订阅者数量
// 2. error: 发布失败时返回的错误
func (c *Redis) Publish(ctx context.Context, channel string, message any) (int64, error) {
	channel = c.formatKey(channel)
	return c.GetRedisCmd(ctx).Publish(ctx, channel, WrapBinaryMarshaler(message)).Result()
}

// PubSubChannels 返回所有订阅频道的列表
// PUBSUB CHANNELS [pattern]
// https://redis.io/commands/pubsub
// 返回值：
// 1. []string: 频道列表
// 2. error: 发布失败时返回的错误
func (c *Redis) PubSubChannels(ctx context.Context, pattern string) ([]string, error) {
	pattern = c.formatKey(pattern)
	return c.GetRedisCmd(ctx).PubSubChannels(ctx, pattern).Result()
}

// PubSubNumSub 返回所有订阅频道的订阅者数量
// PUBSUB NUMSUB [channel-1 ... channel-N]
// https://redis.io/commands/pubsub
// 返回值：
// 1. map[string]int64: 频道->订阅者数量
// 2. error: 发布失败时返回的错误
func (c *Redis) PubSubNumSub(ctx context.Context, channels ...string) (map[string]int64, error) {
	_channels := c.formatKeys(channels)
	return c.GetRedisCmd(ctx).PubSubNumSub(ctx, _channels...).Result()
}

// PubSubNumPat 返回所有订阅模式的数量
// PUBSUB NUMPAT
// https://redis.io/commands/pubsub
// 返回值：
// 1. int64: 订阅模式的数量
// 2. error: 发布失败时返回的错误
func (c *Redis) PubSubNumPat(ctx context.Context) (int64, error) {
	return c.GetRedisCmd(ctx).PubSubNumPat(ctx).Result()
}

// Subscribe 订阅一个或多个频道的信息
// SUBSCRIBE channel [channel ...]
// https://redis.io/commands/subscribe
// 返回值：
// 1. *redis.PubSub: 订阅的频道
func (c *Redis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	_channels := c.formatKeys(channels)
	return c.originalClient.Subscribe(ctx, _channels...)
}

// PSubscribe 订阅一个或多个频道的信息，channel可以使用通配符
// PSUBSCRIBE pattern [pattern ...]
// https://redis.io/commands/psubscribe
func (c *Redis) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	_channels := c.formatKeys(channels)
	return c.originalClient.PSubscribe(ctx, _channels...)
}

// SubscribeHandler 订阅一个或多个频道的信息，使用handler处理接收到的消息
// SUBSCRIBE channel [channel ...]
// https://redis.io/commands/subscribe
func (c *Redis) SubscribeHandler(ctx context.Context, handler func(ctx context.Context, message string), channels ...string) error {
	for m := range c.Subscribe(ctx, channels...).Channel() {
		handler(ctx, m.Payload)
	}
	return redis.ErrClosed
}
