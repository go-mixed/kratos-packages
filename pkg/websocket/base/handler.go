package base

import (
	"context"
)

type IHandler interface {
	OnPong(context.Context, IConnection) error
	OnError(context.Context, IConnection, error)

	OnClose(ctx context.Context, conn IConnection, code int, closeMessage string) error
	OnConnect(context.Context, IConnection) error
	OnDisconnect(context.Context, IConnection) error

	OnRecvMessage(ctx context.Context, conn IConnection, messageType int, message []byte) error
	OnSendMessage(context.Context, IEnvelope, SentConnections) error

	OnStart(context.Context)
	OnStop(context.Context)
}

type IHandleCaller interface {
	CallPongHandler(conn IConnection) error
	CallRecvMessageHandler(conn IConnection, msgType int, msg []byte) error
	CallErrorHandler(conn IConnection, err error)
	CallCloseHandler(conn IConnection, code int, text string) error
}

type UnimplementedWSHandler struct{}

var _ IHandler = (*UnimplementedWSHandler)(nil)

func (s *UnimplementedWSHandler) OnPong(ctx context.Context, conn IConnection) error {
	return nil
}
func (s *UnimplementedWSHandler) OnError(ctx context.Context, conn IConnection, err error) {
}
func (s *UnimplementedWSHandler) OnClose(ctx context.Context, conn IConnection, i int, error string) error {
	return nil
}
func (s *UnimplementedWSHandler) OnConnect(ctx context.Context, conn IConnection) error {
	return nil
}
func (s *UnimplementedWSHandler) OnDisconnect(ctx context.Context, conn IConnection) error {
	return nil
}
func (s *UnimplementedWSHandler) OnRecvMessage(ctx context.Context, conn IConnection, i int, bytes []byte) error {
	return nil
}
func (s *UnimplementedWSHandler) OnSendMessage(ctx context.Context, e IEnvelope, sentSessions SentConnections) error {
	return nil
}
func (s *UnimplementedWSHandler) OnStart(ctx context.Context) {}
func (s *UnimplementedWSHandler) OnStop(ctx context.Context)  {}
