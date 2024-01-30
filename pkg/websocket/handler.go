package websocket

import (
	"context"
)

type IWSHandler interface {
	PongHandler(context.Context, *Session) error
	ErrorHandler(context.Context, *Session, error)

	CloseHandler(context.Context, *Session, int, string) error
	ConnectHandler(context.Context, *Session) error
	DisconnectHandler(context.Context, *Session) error

	RecvMessageHandler(context.Context, *Session, int, []byte) error
	SendMessageHandler(context.Context, IEnvelope, SentSessions) error

	StartHandler(context.Context)
	StopHandler(context.Context)
}

type UnimplementedWSHandler struct{}

var _ IWSHandler = (*UnimplementedWSHandler)(nil)

func (s *UnimplementedWSHandler) PongHandler(ctx context.Context, session *Session) error       { return nil }
func (s *UnimplementedWSHandler) ErrorHandler(ctx context.Context, session *Session, err error) {}
func (s *UnimplementedWSHandler) CloseHandler(ctx context.Context, session *Session, i int, error string) error {
	return nil
}
func (s *UnimplementedWSHandler) ConnectHandler(ctx context.Context, session *Session) error {
	return nil
}
func (s *UnimplementedWSHandler) DisconnectHandler(ctx context.Context, session *Session) error {
	return nil
}
func (s *UnimplementedWSHandler) RecvMessageHandler(ctx context.Context, session *Session, i int, bytes []byte) error {
	return nil
}
func (s *UnimplementedWSHandler) SendMessageHandler(ctx context.Context, e IEnvelope, sentSessions SentSessions) error {
	return nil
}
func (s *UnimplementedWSHandler) StartHandler(ctx context.Context) {}
func (s *UnimplementedWSHandler) StopHandler(ctx context.Context)  {}
