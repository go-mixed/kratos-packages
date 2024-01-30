package worker

import "time"

type onceOption func(*onceWorker)

// WithExpiration 设置过期时间
func WithExpiration(timeRange time.Duration) onceOption {
	return func(w *onceWorker) {
		w.worker.cache = w.worker.cache.WithExpiration(timeRange)
	}
}

func WithKeyPrefix(keyPrefix string) onceOption {
	return func(w *onceWorker) {
		w.worker.cache = w.worker.cache.WithKeyPrefix(keyPrefix)
	}
}
