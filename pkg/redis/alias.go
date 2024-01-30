package redis

import (
	"github.com/redis/go-redis/v9"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
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

var interfacesToStrings = utils.InterfacesToStrings
