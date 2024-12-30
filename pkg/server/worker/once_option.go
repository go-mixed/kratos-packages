package worker

type onceOption func(once *onceWorker)

func WithKeyAsTaskID(val bool) onceOption {
	return func(once *onceWorker) {
		once.keyAsTaskID = val
	}
}
