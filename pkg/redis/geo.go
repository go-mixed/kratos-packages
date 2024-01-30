package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func (c *Redis) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoAdd(ctx, key, geoLocation...).Result()
}

func (c *Redis) GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoPos(ctx, key, members...).Result()
}

func (c *Redis) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoRadius(ctx, key, longitude, latitude, query).Result()
}

func (c *Redis) GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoRadiusStore(ctx, key, longitude, latitude, query).Result()
}

func (c *Redis) GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoRadiusByMember(ctx, key, member, query).Result()
}

func (c *Redis) GeoRadiusByMemberStore(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoRadiusByMemberStore(ctx, key, member, query).Result()
}

func (c *Redis) GeoSearch(ctx context.Context, key string, q *redis.GeoSearchQuery, actual any) ([]string, error) {
	key = c.formatKey(key)
	res := c.GetRedisCmd(ctx).GeoSearch(ctx, key, q)

	err := ScanStringSliceCmd(res.Err(), res.Val(), actual)
	if err != nil {
		return nil, filterNil(err)
	}
	return res.Val(), err
}

func (c *Redis) GeoSearchLocation(ctx context.Context, key string, q *redis.GeoSearchLocationQuery) ([]redis.GeoLocation, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoSearchLocation(ctx, key, q).Result()
}

func (c *Redis) GeoSearchStore(ctx context.Context, key, store string, q *redis.GeoSearchStoreQuery) (int64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoSearchStore(ctx, key, store, q).Result()
}

func (c *Redis) GeoDist(ctx context.Context, key string, member1, member2, unit string) (float64, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoDist(ctx, key, member1, member2, unit).Result()
}

func (c *Redis) GeoHash(ctx context.Context, key string, members ...string) ([]string, error) {
	key = c.formatKey(key)
	return c.GetRedisCmd(ctx).GeoHash(ctx, key, members...).Result()
}
