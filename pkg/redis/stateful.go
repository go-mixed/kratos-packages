package redis

import (
	"context"
)

// Auth 验证密码
// AUTH password
// https://redis.io/commands/auth
func (c *Redis) Auth(ctx context.Context, password string) (string, error) {
	return c.originalClient.Conn().Auth(ctx, password).Result()
}

// AuthACL 验证密码
// AUTH username password
// https://redis.io/commands/auth
func (c *Redis) AuthACL(ctx context.Context, username, password string) (string, error) {
	return c.originalClient.Conn().AuthACL(ctx, username, password).Result()
}

// Select 切换数据库，0-15
// SELECT index
// https://redis.io/commands/select
func (c *Redis) Select(ctx context.Context, index int) (string, error) {
	return c.originalClient.Conn().Select(ctx, index).Result()
}

// SwapDB 交换两个数据库的数据
// SWAPDB index1 index2
// https://redis.io/commands/swapdb
func (c *Redis) SwapDB(ctx context.Context, index1, index2 int) (string, error) {
	return c.originalClient.Conn().SwapDB(ctx, index1, index2).Result()
}

// ClientSetName 设置客户端名称
// CLIENT SETNAME connection-name
// https://redis.io/commands/client-setname
func (c *Redis) ClientSetName(ctx context.Context, name string) (bool, error) {
	return c.originalClient.Conn().ClientSetName(ctx, name).Result()
}
