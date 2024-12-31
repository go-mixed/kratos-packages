package worker

type onceOption func(once *onceWorker)

// WithOverwrite 是否覆盖同名的taskId的任务，默认不覆盖
func WithOverwrite(val bool) onceOption {
	return func(once *onceWorker) {
		once.overwrite = val
	}
}
