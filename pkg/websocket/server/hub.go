package server

import (
	"context"
	"github.com/gorilla/websocket"
	"go.uber.org/multierr"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/envelope"
)

type Hub struct {
	app    *app.App
	logger *log.Helper
	server *Server
}

func NewHub(
	app *app.App,
	logger log.Logger,
	server *Server,
) base.IHub {
	logHelper := log.NewModuleHelper(logger, "websocket/hub")

	return &Hub{
		app:    app,
		logger: logHelper,
		server: server,
	}
}

func (h *Hub) RegisterHandlers(handlers ...base.IHandler) {
	h.server.handlers = append(h.server.handlers, handlers...)
}

func (h *Hub) Running() bool {
	return h.server.Running()
}

func (h *Hub) GetConnections() base.IConnections {
	return h.server.connections
}

func (h *Hub) GetAllConnectionIDs(ctx context.Context) []base.ConnectionID {
	return h.server.connections.GetAllConnectionIDs(ctx)
}

func (h *Hub) Send(ctx context.Context, envelope base.IEnvelope) error {
	// 先执行当前节点的发送
	err := h.server.sendRaw(ctx, envelope)
	// 发送到集群
	err = multierr.Append(err, h.server.cluster.Publish(ctx, envelope))
	return err
}

func (h *Hub) SendText(ctx context.Context, message string, connIds ...base.ConnectionID) error {
	return h.Send(ctx, envelope.NewEnvelopeBuilder().
		WithAppID(h.app.ID()).
		WithContext(ctx).
		WithMessageType(base.TextMessage).
		WithMessage([]byte(message)).
		WithConnectionIDs(false, connIds...).
		Build())
}

func (h *Hub) SendBinary(ctx context.Context, message []byte, connIds ...base.ConnectionID) error {
	return h.Send(ctx, envelope.NewEnvelopeBuilder().
		WithAppID(h.app.ID()).
		WithContext(ctx).
		WithMessageType(base.BinaryMessage).
		WithMessage(message).
		WithConnectionIDs(false, connIds...).
		Build())
}

func (h *Hub) BroadcastText(ctx context.Context, message string) error {
	return h.Send(ctx, envelope.NewEnvelopeBuilder().
		WithAppID(h.app.ID()).
		WithContext(ctx).
		WithMessageType(base.TextMessage).
		WithMessage([]byte(message)).
		WithConnectionIDs(true).
		Build())
}

func (h *Hub) BroadcastBinary(ctx context.Context, message []byte) error {
	return h.Send(ctx, envelope.NewEnvelopeBuilder().
		WithAppID(h.app.ID()).
		WithContext(ctx).
		WithMessageType(base.BinaryMessage).
		WithMessage(message).
		WithConnectionIDs(true).
		Build())
}

func (h *Hub) Close(ctx context.Context, exitMessage string, connIds ...base.ConnectionID) {
	// 先关闭本地节点
	conns := h.server.connections.MGet(connIds...)
	for _, conn := range conns.Iterator() {
		_ = conn.TryClose(exitMessage)
	}

	// 发送关闭消息到集群
	_envelope := envelope.NewEnvelopeBuilder().
		WithAppID(h.app.ID()).
		WithContext(ctx).
		WithMessageType(base.CloseMessage).
		WithMessage(websocket.FormatCloseMessage(websocket.CloseNormalClosure, exitMessage)).
		WithConnectionIDs(false, connIds...).
		Build()

	_ = h.server.cluster.Publish(ctx, _envelope)
}
