package worker

type onceOption func(once *onceWorker)

// WithOverride 如果taskId相同，是否覆盖之前的任务，默认不覆盖
func WithOverride(val bool) onceOption {
	return func(once *onceWorker) {
		once.override = val
	}
}
