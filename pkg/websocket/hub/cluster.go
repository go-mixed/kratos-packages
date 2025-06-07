package hub

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/envelope"
	"strings"
	"time"
)

type wsCluster struct {
	logger  *log.Helper
	rdb     *redis.Client
	channel string

	broadcaster redis.Broadcaster

	app *app.App

	sendFunc func(ctx context.Context, envelope base.IEnvelope) error
}

// newWsCluster 创建集群广播
//
//	 注意：
//		keyPrefix: 不同的环境需要设置不同的前缀。因为PUB/SUB是redis全局的，无视redis的db设置
//		详见：https://redis.io/docs/interact/pubsub/#database--scoping
func newWsCluster(
	app *app.App,
	logger *log.Helper,
	rdb *redis.Redis,
) *wsCluster {

	return &wsCluster{
		logger:  logger,
		app:     app,
		channel: "websocket_broadcast",

		broadcaster: rdb,
	}
}

func (s *wsCluster) handleSubscribe(ctx context.Context, message string) {
	_envelope, err := envelope.UnmarshalIEnvelope([]byte(message))
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[Cluster]unmarshal message error: %v", err)
		return
	}

	// 排除自己
	if strings.EqualFold(_envelope.GetOriginalAppID(), s.app.ID()) {
		return
	}
	// 设置第一次发送者的trace
	newCtx := _envelope.GetContext(ctx)

	if err = s.sendFunc(newCtx, _envelope); err != nil {
		s.logger.WithContext(newCtx).Errorf("send message error: %v", err)
	}
}

// Publish 广播消息到集群（排除自己）
func (s *wsCluster) Publish(ctx context.Context, _envelope base.IEnvelope) error {
	// Publish 广播消息到集群（排除自己）
	_envelope.SetOriginalAppID(s.app.ID())
	j, err := envelope.MarshalIEnvelope(_envelope)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[Cluster]marshal message error: %v", err)
	}

	_, err = s.broadcaster.Publish(ctx, s.channel, j)
	return err
}

// Listen 启动监听集群广播
func (s *wsCluster) Listen(ctx context.Context, sendFunc func(context.Context, base.IEnvelope) error) {
	s.sendFunc = sendFunc
	go s.doSubscribe(ctx)
}

// doSubscribe 当连接断开之后，尝试重连
func (s *wsCluster) doSubscribe(ctx context.Context) {
	for {
		if err := s.broadcaster.SubscribeHandler(ctx, s.handleSubscribe, s.channel); err != nil {
			s.logger.WithContext(ctx).Errorf("subscribe error, retrying for 10s: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(10 * time.Second)
		}
	}
}
