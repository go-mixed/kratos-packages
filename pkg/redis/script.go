package redis

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"github.com/redis/go-redis/v9"
	"io"
	"strings"
)

type chainScript struct {
	src, hash string
	redis     *Redis
}

func newScript(redis *Redis, src string) *chainScript {
	h := sha1.New()
	_, _ = io.WriteString(h, src)
	return &chainScript{
		src:   src,
		hash:  hex.EncodeToString(h.Sum(nil)),
		redis: redis,
	}
}

func (s *chainScript) Load(ctx context.Context) (string, error) {
	return s.redis.ScriptLoad(ctx, s.src)
}

func (s *chainScript) Exists(ctx context.Context) ([]bool, error) {
	return s.redis.ScriptExists(ctx)
}

func (s *chainScript) Eval(ctx context.Context, keys []string, args ...any) *redis.Cmd {
	return s.redis.Eval(ctx, s.src, keys, args...)
}

func (s *chainScript) EvalSha(ctx context.Context, keys []string, args ...any) *redis.Cmd {
	return s.redis.EvalSha(ctx, s.hash, keys, args...)
}

// Run optimistically uses EVALSHA to run the script. If script does not exist
// it is retried using EVAL.
func (s *chainScript) Run(ctx context.Context, keys []string, args ...any) *redis.Cmd {
	r := s.EvalSha(ctx, keys, args...)
	if err := r.Err(); err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		return s.Eval(ctx, keys, args...)
	}
	return r
}
