package worker

import "time"

type onceOption func(*onceWorker)

// WithExpiration 设置过期时间
func WithExpiration(timeout time.Duration) onceOption {
	return func(w *onceWorker) {
		redisOptions := w.worker.options.redisOptions
		redisOptions.Expiration = timeout
		w.worker.store = w.worker.store.WithOptions(redisOptions)
	}
}

func WithKeyPrefix(keyPrefix string) onceOption {
	return func(w *onceWorker) {
		redisOptions := w.worker.options.redisOptions
		redisOptions.KeyPrefix = keyPrefix
		w.worker.store = w.worker.store.WithOptions(redisOptions)
	}
}
