package stream

import "time"

type options struct {
	flushInterval time.Duration
	flushSize     int32
	contentType   string
}

type StreamOption func(*options)

// WithFlushInterval 设置刷新间隔时间。
func WithFlushInterval(interval time.Duration) StreamOption {
	return func(o *options) {
		o.flushInterval = interval
	}
}

// WithFlushSize 设置刷新大小。
func WithFlushSize(size int) StreamOption {
	return func(o *options) {
		o.flushSize = int32(size)
	}
}
