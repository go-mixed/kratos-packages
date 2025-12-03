package queue

import (
	"time"

	"github.com/hdt3213/delayqueue"
)

type opt func(*DelayQueue)

// WithCallback set callback for queue to receives and consumes messages
// callback returns true to confirm successfully consumed, false to re-deliver this message
// 为当前队列设置callback函数，并接受消息
// 当callback返回true时，表示消息已经成功消费，否则，会重新投递这个消息。
// 如果 WithMaxConsumeDuration 时间内没返回，也会重新投递。直到重试次数超过 WithDefaultRetryCount。且重新投递的延时受 WithNackRedeliveryDelay 控制
func WithCallback(callback delayqueue.CallbackFunc) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithCallback(callback)
	}
}

// WithLogger customizes logger for queue
// 自定义日志记录器。
func WithLogger(logger delayqueue.Logger) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithLogger(logger)
	}
}

// WithFetchInterval customizes the interval at which consumer fetch message from redis
// 自定义从redis中拉取消息的时间间隔。这个意思就是，消息并不是实时拉取的，而是按照这个时间间隔拉取的。（即使delay设置为0，也会按照这个时间间隔拉取消息）
func WithFetchInterval(d time.Duration) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithFetchInterval(d)
	}
}

// WithScriptPreload use script load command preload scripts to redis
// 使用redis的script load命令预加载脚本到redis
func WithScriptPreload(flag bool) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithScriptPreload(flag)
	}
}

// WithMaxConsumeDuration customizes max consume duration
// If no acknowledge received within WithMaxConsumeDuration after message delivery, DelayQueue will try to deliver this message again
// 最大消费时间，超过这个时间后，消息会被重新投递。
// 即：如果没有在 WithMaxConsumeDuration 时间内收到确认（即回调函数返回true），消息会被重新投递。
func WithMaxConsumeDuration(d time.Duration) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithMaxConsumeDuration(d)
	}
}

// WithFetchLimit limits the max number of processing messages, 0 means no limit
// 每次从队列中拉取的消息数量。
func WithFetchLimit(limit uint) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithFetchLimit(limit)
	}
}

// WithConcurrent sets the number of concurrent consumers.
// 并发消费的数量。
func WithConcurrent(c uint) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithConcurrent(c)
	}
}

// WithDefaultRetryCount customizes the max number of retry, it effects of messages in this queue
// use WithRetryCount during DelayQueue.SendScheduleMsg or DelayQueue.SendDelayMsg to specific retry count of particular message.
// 最大重试次数，它会影响队列中的所有消息。
// 在 DelayQueue.SendScheduleMsg 或 DelayQueue.SendDelayMsg 中，可以通过 WithRetryCount 来指定特定消息的重试次数。
func WithDefaultRetryCount(count uint) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithDefaultRetryCount(count)
	}
}

// WithNackRedeliveryDelay customizes the interval between redelivery and nack (callback returns false)
// If consumption exceeded deadline, the message will be redelivered immediately
// 消息处理失败后（nack或return false）的重新投递延迟时间。
// 即：当消息处理失败（返回 false 或 NACK）后，系统会等待指定的时间间隔再重新尝试投递该消息。
// 如果消费超过了 WithNackRedeliveryDelay ，消息会立即重新投递。
func WithNackRedeliveryDelay(d time.Duration) opt {
	return func(queue *DelayQueue) {
		queue.DelayQueue = queue.DelayQueue.WithNackRedeliveryDelay(d)
	}
}
