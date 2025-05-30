package base

import (
	"context"
)

type IHandler interface {
	PongHandler(context.Context, IConnection) error
	ErrorHandler(context.Context, IConnection, error)

	CloseHandler(context.Context, IConnection, int, string) error
	ConnectHandler(context.Context, IConnection) error
	DisconnectHandler(context.Context, IConnection) error

	RecvMessageHandler(context.Context, IConnection, int, []byte) error
	SendMessageHandler(context.Context, IEnvelope, SentConnections) error

	StartHandler(context.Context)
	StopHandler(context.Context)
}

type IHandleCaller interface {
	CallPongHandler(conn IConnection) error
	CallRecvMessageHandler(conn IConnection, msgType int, msg []byte) error
	CallSendMessageHandler(ctx context.Context, envelope IEnvelope, sentSessions SentConnections) error
	CallErrorHandler(conn IConnection, err error)
	CallCloseHandler(conn IConnection, code int, text string) error
}

type UnimplementedWSHandler struct{}

var _ IHandler = (*UnimplementedWSHandler)(nil)

func (s *UnimplementedWSHandler) PongHandler(ctx context.Context, conn IConnection) error {
	return nil
}
func (s *UnimplementedWSHandler) ErrorHandler(ctx context.Context, conn IConnection, err error) {
}
func (s *UnimplementedWSHandler) CloseHandler(ctx context.Context, conn IConnection, i int, error string) error {
	return nil
}
func (s *UnimplementedWSHandler) ConnectHandler(ctx context.Context, conn IConnection) error {
	return nil
}
func (s *UnimplementedWSHandler) DisconnectHandler(ctx context.Context, conn IConnection) error {
	return nil
}
func (s *UnimplementedWSHandler) RecvMessageHandler(ctx context.Context, conn IConnection, i int, bytes []byte) error {
	return nil
}
func (s *UnimplementedWSHandler) SendMessageHandler(ctx context.Context, e IEnvelope, sentSessions SentConnections) error {
	return nil
}
func (s *UnimplementedWSHandler) StartHandler(ctx context.Context) {}
func (s *UnimplementedWSHandler) StopHandler(ctx context.Context)  {}
