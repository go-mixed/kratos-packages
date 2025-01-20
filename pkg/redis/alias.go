package redis

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	
)

var (
	ErrClosed = redis.ErrClosed
)

type (
	Client   = redis.Client
	PubSub   = redis.PubSub
	Z        = redis.Z
	ZRangeBy = redis.ZRangeBy
	Cmdable  = redis.Cmdable
)

const Nil = redis.Nil

var InstrumentTracing = redisotel.InstrumentTracing
