package worker

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"runtime"
)

type workerOptions struct {
	redisOptions redis.Options
	workerCount  int
	wheelSize    int
}

func DefaultWorkerOptions() workerOptions {
	return workerOptions{
		redisOptions: redis.DefaultOptions(),
		workerCount:  runtime.NumCPU() * 2,
		wheelSize:    20,
	}
}

type workerOption func(options *workerOptions)

// WithRedisOptions redis配置
func WithRedisOptions(options redis.Options) workerOption {
	return func(opts *workerOptions) {
		opts.redisOptions = options
	}
}

// WithWorkerCount 工作者最大数量，默认为cpu * 2
func WithWorkerCount(count int) workerOption {
	return func(opts *workerOptions) {
		opts.workerCount = count
	}
}

// WithWheelSize 时间轮大小，默认为20
func WithWheelSize(size int) workerOption {
	return func(opts *workerOptions) {
		opts.wheelSize = size
	}
}
