package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Pipeliner struct {
	originalPipeliner redis.Pipeliner
}

// NewPipeliner wraps a Pipeliner with CmdOptions and returns a new Pipeliner.
func NewPipeliner(originalPipeliner redis.Pipeliner) *Pipeliner {
	return &Pipeliner{
		originalPipeliner: originalPipeliner,
	}
}

func (k *Pipeliner) Len() int {
	return k.originalPipeliner.Len()
}

func (k *Pipeliner) Do(ctx context.Context, args ...any) *redis.Cmd {
	return k.originalPipeliner.Do(ctx, args...)
}

func (k *Pipeliner) Process(ctx context.Context, cmd redis.Cmder) error {
	return k.originalPipeliner.Process(ctx, cmd)
}

func (k *Pipeliner) Discard() {
	k.originalPipeliner.Discard()
}

func (k *Pipeliner) Exec(ctx context.Context) ([]redis.Cmder, error) {
	return k.originalPipeliner.Exec(ctx)
}
