package worker

import "gopkg.in/go-mixed/kratos-packages.v2/pkg/server/task"

type onceOption func(once *onceWorker)

// WithTaskId 提交的任务，返回这个taskId，默认是"once:cron:key"、"once:timer:key"、"once:immediate:key"这种格式
func WithTaskId(taskId task.TaskID) onceOption {
	return func(once *onceWorker) {
		once.taskId = taskId
	}
}

// WithOverride 如果taskId相同，是否覆盖之前的任务，默认不覆盖
func WithOverride(val bool) onceOption {
	return func(once *onceWorker) {
		once.override = val
	}
}
