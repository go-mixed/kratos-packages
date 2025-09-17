package queue

import (
	"context"
	"sync/atomic"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/redis/go-redis/v9"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
)
import "github.com/hdt3213/delayqueue"

type QueueServer struct {
	queueTable  utils.ConcurrentMap[string, *delayqueue.DelayQueue]
	redisClient *redis.Client

	kratosStarted atomic.Bool
}

var _ transport.Server = (*QueueServer)(nil)

func NewQueueServer(rdb *redis.Client) *QueueServer {
	return &QueueServer{
		redisClient:   rdb,
		queueTable:    utils.ConcurrentMap[string, *delayqueue.DelayQueue]{},
		kratosStarted: atomic.Bool{},
	}
}

// NewQueue 新建队列，如果kratos已经启动，则立即启动消费该队列；否则，会在kratos启动时统一启动消费
func (q *QueueServer) NewQueue(name string, callback delayqueue.CallbackFunc, opts ...opt) *delayqueue.DelayQueue {
	queue := delayqueue.NewQueue(name, q.redisClient, callback)

	for _, _opt := range opts {
		_opt(queue)
	}

	// kratos已经启动，说明不是在构造函数中新建的，则显式启动该队列
	// 否则，会在kratos启动时统一启动
	if q.kratosStarted.Load() {
		queue.StartConsume()
	}

	// 必须放在kratosStarted之后，避免切换时（即临界区）新增的queue
	q.queueTable.Store(name, queue)

	return queue
}

// RemoveQueue 删除队列
func (q *QueueServer) RemoveQueue(name string) {
	queue, ok := q.queueTable.LoadAndDelete(name)
	if ok {
		queue.StopConsume()
	}
}

// GetQueue  获取队列
func (q *QueueServer) GetQueue(name string) *delayqueue.DelayQueue {
	queue, _ := q.queueTable.Load(name)
	return queue
}

// HasQueue  是否存在队列
func (q *QueueServer) HasQueue(name string) bool {
	_, ok := q.queueTable.Load(name)
	return ok
}

func (q *QueueServer) Start(ctx context.Context) error {
	q.kratosStarted.Store(true)
	// 此时读取到的队列列表，不包含kratosStarted切换时（即临界区）新增的queue
	// 所以在kratosStarted设置之后新增的queue，会立即在NewQueue中启动消费
	for _, queue := range q.queueTable.Iterator() {
		queue.StartConsume()
	}
	return nil
}

func (q *QueueServer) Stop(ctx context.Context) error {
	// Kratos停止时，会调用Stop
	for _, queue := range q.queueTable.Iterator() {
		queue.StopConsume()
	}
	q.kratosStarted.Store(false)
	return nil
}
