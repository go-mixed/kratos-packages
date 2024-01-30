package redis

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"time"
)

func (c *Redis) Command(ctx context.Context) *redis.CommandsInfoCmd {
	return c.GetRedisCmd(ctx).Command(ctx)
}

func (c *Redis) ClientGetName(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ClientGetName(ctx).Result()
}

func (c *Redis) Echo(ctx context.Context, message any) (string, error) {
	return c.GetRedisCmd(ctx).Echo(ctx, message).Result()
}

func (c *Redis) Ping(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).Ping(ctx).Result()
}

func (c *Redis) Quit(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).Quit(ctx).Result()
}

func (c *Redis) BgRewriteAOF(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).BgRewriteAOF(ctx).Result()
}

func (c *Redis) BgSave(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).BgSave(ctx).Result()
}

func (c *Redis) ClientKill(ctx context.Context, ipPort string) (string, error) {
	return c.GetRedisCmd(ctx).ClientKill(ctx, ipPort).Result()
}

func (c *Redis) ClientKillByFilter(ctx context.Context, keys ...string) (int64, error) {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).ClientKillByFilter(ctx, _keys...).Result()
}

func (c *Redis) ClientList(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ClientList(ctx).Result()
}

func (c *Redis) ClientPause(ctx context.Context, dur time.Duration) (bool, error) {
	return c.GetRedisCmd(ctx).ClientPause(ctx, dur).Result()
}

func (c *Redis) ClientID(ctx context.Context) (int64, error) {
	return c.GetRedisCmd(ctx).ClientID(ctx).Result()
}

// ConfigGet 获取配置
// https://redis.io/commands/config-get
// actual的结构是&struct{A int `redis:"a"`, ...}，并且无法内嵌Struct、Map，功能十分有限
func (c *Redis) ConfigGet(ctx context.Context, parameter string, actual any) ([]any, error) {
	res := c.GetRedisCmd(ctx).ConfigGet(ctx, parameter)

	err := utils.IfFunc(actual != nil, func() error { return res.Scan(actual) }, func() error { return res.Err() })
	if err != nil {
		return nil, filterNil(err)
	}

	return res.Val(), nil
}

func (c *Redis) ConfigResetStat(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ConfigResetStat(ctx).Result()
}

func (c *Redis) ConfigSet(ctx context.Context, parameter, value string) (string, error) {
	return c.GetRedisCmd(ctx).ConfigSet(ctx, parameter, value).Result()
}

func (c *Redis) ConfigRewrite(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ConfigRewrite(ctx).Result()
}

func (c *Redis) DBSize(ctx context.Context) (int64, error) {
	return c.GetRedisCmd(ctx).DBSize(ctx).Result()
}

func (c *Redis) FlushAll(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).FlushAll(ctx).Result()
}

func (c *Redis) FlushAllAsync(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).FlushAllAsync(ctx).Result()
}

func (c *Redis) FlushDB(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).FlushDB(ctx).Result()
}

func (c *Redis) FlushDBAsync(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).FlushDBAsync(ctx).Result()
}

func (c *Redis) Info(ctx context.Context, section ...string) (string, error) {
	return c.GetRedisCmd(ctx).Info(ctx, section...).Result()
}

func (c *Redis) LastSave(ctx context.Context) (int64, error) {
	return c.GetRedisCmd(ctx).LastSave(ctx).Result()
}

func (c *Redis) Save(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).Save(ctx).Result()
}

func (c *Redis) Shutdown(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).Shutdown(ctx).Result()
}

func (c *Redis) ShutdownSave(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ShutdownSave(ctx).Result()
}

func (c *Redis) ShutdownNoSave(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ShutdownNoSave(ctx).Result()
}

func (c *Redis) SlaveOf(ctx context.Context, host, port string) (string, error) {
	return c.GetRedisCmd(ctx).SlaveOf(ctx, host, port).Result()
}

func (c *Redis) Time(ctx context.Context) (time.Time, error) {
	return c.GetRedisCmd(ctx).Time(ctx).Result()
}

func (c *Redis) ReadOnly(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ReadOnly(ctx).Result()
}

func (c *Redis) ReadWrite(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ReadWrite(ctx).Result()
}
