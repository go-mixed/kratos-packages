package queue

import (
	"time"

	"github.com/hdt3213/delayqueue"
)

type opt func(*delayqueue.DelayQueue) *delayqueue.DelayQueue

func WithCallback(callback delayqueue.CallbackFunc) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithCallback(callback)
	}
}

func WithLogger(logger delayqueue.Logger) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithLogger(logger)
	}
}

func WithFetchInterval(d time.Duration) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithFetchInterval(d)
	}
}

func WithScriptPreload(flag bool) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithScriptPreload(flag)
	}
}

func WithMaxConsumeDuration(d time.Duration) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithMaxConsumeDuration(d)
	}
}

func WithFetchLimit(limit uint) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithFetchLimit(limit)
	}
}

func WithConcurrent(c uint) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithConcurrent(c)
	}
}

func WithDefaultRetryCount(count uint) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithDefaultRetryCount(count)
	}
}

func WithNackRedeliveryDelay(d time.Duration) opt {
	return func(queue *delayqueue.DelayQueue) *delayqueue.DelayQueue {
		return queue.WithNackRedeliveryDelay(d)
	}
}

//func (q *QueueServer) WithCallback(callback delayqueue.CallbackFunc) opt {
//	return WithCallback(callback)
//}
//
//func (q *QueueServer) WithLogger(logger delayqueue.Logger) opt {
//	return WithLogger(logger)
//}
//
//func (q *QueueServer) WithFetchInterval(d time.Duration) opt {
//	return WithFetchInterval(d)
//}
//
//func (q *QueueServer) WithScriptPreload(flag bool) opt {
//	return WithScriptPreload(flag)
//}
//
//func (q *QueueServer) WithMaxConsumeDuration(d time.Duration) opt {
//	return WithMaxConsumeDuration(d)
//}
//
//func (q *QueueServer) WithFetchLimit(limit uint) opt {
//	return WithFetchLimit(limit)
//}
//
//func (q *QueueServer) WithConcurrent(c uint) opt {
//	return WithConcurrent(c)
//}
//
//func (q *QueueServer) WithDefaultRetryCount(count uint) opt {
//	return WithDefaultRetryCount(count)
//}
//
//func (q *QueueServer) WithNackRedeliveryDelay(d time.Duration) opt {
//	return WithNackRedeliveryDelay(d)
//}
