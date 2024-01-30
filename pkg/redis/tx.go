package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Tx struct {
	originalClient *Redis
	originalTx     *redis.Tx
	options        Options
}

func NewTx(
	oldClient *Redis,
	originalTx *redis.Tx,
	options Options) *Tx {
	return &Tx{
		originalClient: oldClient,
		originalTx:     originalTx,
		options:        options,
	}
}

func (t *Tx) Context() context.Context {
	return t.originalTx.Context()
}

func (t *Tx) WithContext(ctx context.Context) *Tx {
	return &Tx{
		originalTx: t.originalTx.WithContext(ctx),
		options:    t.options,
	}
}

func (t *Tx) Process(ctx context.Context, cmd redis.Cmder) error {
	return t.originalTx.Process(ctx, cmd)
}

func (t *Tx) Close(ctx context.Context) error {
	return t.originalTx.Close(ctx)
}

func (t *Tx) Watch(ctx context.Context, keys ...string) (string, error) {
	_keys := t.originalClient.formatKeys(keys)
	return t.originalTx.Watch(ctx, _keys...).Result()
}

func (t *Tx) Unwatch(ctx context.Context, keys ...string) (string, error) {
	_keys := t.originalClient.formatKeys(keys)
	return t.originalTx.Unwatch(ctx, _keys...).Result()
}

func (t *Tx) Pipelined(ctx context.Context, fns ...func(ctx context.Context) error) ([]redis.Cmder, error) {
	return t.originalTx.Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
		ctx = newContext(ctx, pipeliner)
		for _, fn := range fns {
			err := fn(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *Tx) TxPipelined(ctx context.Context, fns ...func(ctx context.Context) error) ([]redis.Cmder, error) {
	return t.originalTx.TxPipelined(ctx, func(pipeliner redis.Pipeliner) error {
		ctx = newContext(ctx, pipeliner)
		for _, fn := range fns {
			err := fn(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
