package queue

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/hdt3213/delayqueue"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
)

type DelayQueue struct {
	*delayqueue.DelayQueue
	name string
	rdb  *redis.Client

	handlers map[string]reflect.Value
	logger   *log.Helper
	app      *app.App
}

// CallbackFunc 是延迟队列的回调函数类型
// 它接收一个上下文和一个参数，返回一个错误
// 如果返回 nil，则表示执行成功；否则表示执行失败，会打印错误日志，并且重新放入队列，等待 WithNackRedeliveryDelay 之后重试执行，直到超过 WithDefaultRetryCount 次
type CallbackFunc[Arg any] func(ctx context.Context, arg Arg) error

type taskEnvelope struct {
	TypeName string
	Once     bool
	Arg      any
}

// NewDelayQueue 创建一个新的延迟队列实例
func NewDelayQueue(name string, rdb *redis.Client, logger *log.Helper) *DelayQueue {
	queue := &DelayQueue{name: name, rdb: rdb, handlers: make(map[string]reflect.Value), logger: logger}
	queue.DelayQueue = delayqueue.NewQueue(name, rdb).WithLogger(logger).WithCallback(queue.callback)
	return queue
}

// callback 是延迟队列的回调函数，用于处理延迟消息
// 它会根据消息的 TypeName 查找对应的回调函数并执行
// 如果执行成功，返回 true；如果执行失败，返回 false，然后会重试执行
func (q *DelayQueue) callback(payload string) (ack bool) {
	var h taskEnvelope
	if err := gob.NewDecoder(strings.NewReader(payload)).Decode(&h); err != nil {
		return true // 解码失败，直接返回 true 不处理
	}

	ctx := lo.IfF(q.app != nil, func() context.Context {
		return q.app.BaseContext()
	}).Else(context.Background())

	if h.Once {
		// 检查是否已经执行过。互斥有效期是第一次成功任务执行之后的24小时内
		if ok, err := q.rdb.SetNX(ctx, fmt.Sprintf("queue:%s:once:%x", q.name, utils.MD5String(payload)), time.Now().UnixNano(), 24*time.Hour).Result(); !ok && err == nil {
			return true // 如果已经执行过，直接返回 true 不处理
		}
	}

	// 查找对应的回调函数
	callback, ok := q.handlers[h.TypeName]
	if !ok {
		q.logger.WithContext(ctx).Warnf("[Queue]no callback found for type %s", h.TypeName)
		return true // 没有找到对应的回调函数，直接返回 true 不处理
	}

	defer func() {
		if err := recover(); err != nil {
			stackTrace := debug.Stack()
			q.logger.WithContext(ctx).Errorf("[Queue]callback %s panic error = %v, stack = %s", h.TypeName, err, string(stackTrace))
		}
	}()

	// 调用回调函数
	results := callback.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(h.Arg)})
	if err, ok := results[0].Interface().(error); ok && err != nil {
		q.logger.WithContext(ctx).Errorf("[Queue]callback %s failed: %v", h.TypeName, err)
		return false // 回调函数执行失败，返回 false 表示需要重试
	}

	return true
}

// sendMessageAt 发送一个指定时间的消息
// 如果 once 为 true，则表示该消息在24小时内只能执行一次
func (q *DelayQueue) sendMessageAt(arg any, at time.Time, once bool) error {
	var h taskEnvelope
	h.TypeName = utils.GetClassName(arg)
	h.Arg = arg
	h.Once = once

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(h); err != nil {
		return err
	}
	return q.DelayQueue.SendScheduleMsg(buf.String(), at)
}

// SendDelayedMessage 发送一个延迟消息，延迟时间为 delay
func (q *DelayQueue) SendDelayedMessage(arg any, delay time.Duration) error {
	return q.sendMessageAt(arg, time.Now().Add(delay), false)
}

// SenMessageAt 发送一个指定时间的消息
func (q *DelayQueue) SenMessageAt(arg any, at time.Time) error {
	return q.sendMessageAt(arg, at, false)
}

// SendOnceDelayedMessage 发送一个延迟消息，延迟时间为 delay。相同的arg参数，不论结果成功失败，只能执行一次（24小时内）。
// 使用的是redis的SetNX作为执行一次的判断依据。为了避免redis中残留大量无效key，该key在第一次任务执行之后的24小时后失效（注意：不是添加任务的24小时后）
// 所以任务执行完24小时后，如果还有相同的arg参数的任务，会成功执行
func (q *DelayQueue) SendOnceDelayedMessage(arg any, delay time.Duration) error {
	return q.sendMessageAt(arg, time.Now().Add(delay), true)
}

// SendOnceMessageAt 发送一个指定时间的消息，相同的arg参数，不论结果成功失败，只能执行一次（24小时内）。
// 使用的是redis的SetNX作为执行一次的判断依据。为了避免redis中残留大量无效key，该key在第一次任务执行之后的24小时后失效（注意：不是添加任务的24小时后）
// 所以任务执行完24小时后，如果还有相同的arg参数的任务，会成功执行
func (q *DelayQueue) SendOnceMessageAt(arg any, at time.Time) error {
	return q.sendMessageAt(arg, at, true)
}

// clone 输出一个新的 DelayQueue 实例
func (q *DelayQueue) clone() *DelayQueue {
	return &DelayQueue{
		DelayQueue: q.DelayQueue,
		name:       q.name,
		rdb:        q.rdb,
		handlers:   q.handlers,
		logger:     q.logger,
	}
}

// RegisterHandler 注册一个回调函数，用于处理指定类型的消息。
// 注意：一个 Arg 类型只能注册一个回调函数。
// 请在 delayqueue.StartConsume 之前注册，不然会有历史消息因为没有handler导致丢失
func RegisterHandler[Arg any](queue *DelayQueue, callback CallbackFunc[Arg]) *DelayQueue {
	var arg Arg
	queue.handlers[utils.GetClassName(arg)] = reflect.ValueOf(callback)
	// 将类型注册到 gob 中，以便编码和解码
	gob.Register(arg)
	return queue
}

// UnregisterHandler 注销一个回调函数
func UnregisterHandler[Arg any](queue *DelayQueue, callback CallbackFunc[Arg]) *DelayQueue {
	var arg Arg
	delete(queue.handlers, utils.GetClassName(arg))
	return queue
}
