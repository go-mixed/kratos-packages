package redis

import (
	"github.com/redis/go-redis/v9"
	"time"
)

type Options struct {
	Expiration          time.Duration
	KeyPrefix           string
	SaveEmptyOnRemember bool
}

func DefaultOptions() Options {
	return Options{
		Expiration:          redis.KeepTTL,
		KeyPrefix:           "",
		SaveEmptyOnRemember: false,
	}
}

func (o Options) WithExpiration(expiration time.Duration) Options {
	o.Expiration = expiration
	return o
}

func (o Options) WithKeyPrefix(keyPrefix string) Options {
	o.KeyPrefix = keyPrefix
	return o
}

func (o Options) WithSaveEmptyOnRemember(saveEmptyOnRemember bool) Options {
	o.SaveEmptyOnRemember = saveEmptyOnRemember
	return o
}
