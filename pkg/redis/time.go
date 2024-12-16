package redis

import (
	"context"
	"time"
)

// ServerTimeDelta 获取redis服务器时间与本地时间的差值（注意：会有socket传输时间的误差）
// e.g.: time.Now().Add(ServerTimeDelta()) 可以得到redis服务器的时间
func (c *Redis) ServerTimeDelta(ctx context.Context) time.Duration {
	serverTime, err := c.Time(ctx)
	if err != nil || serverTime.IsZero() {
		return 0
	}

	return serverTime.Sub(time.Now())
}
