package base

import (
	"context"
	"google.golang.org/protobuf/proto"
)

type connectionKey struct {
}

type IStream interface {
	MessageType() int
	Context() context.Context
	ConnectionID() ConnectionID
	Connection() IConnection
	Send(message []byte) error
	SendProtoMessage(proto.Message) error
}

func NewContext(ctx context.Context, stream IStream) context.Context {
	return context.WithValue(ctx, connectionKey{}, stream)
}

func FromContext(ctx context.Context) (IStream, bool) {
	impl, ok := ctx.Value(connectionKey{}).(IStream)
	return impl, ok
}

type stream struct {
	ctx         context.Context
	hub         IHub
	connection  IConnection
	messageType int
}

func NewStream(ctx context.Context, hub IHub, connection IConnection, messageType int) *stream {
	return &stream{
		ctx:         ctx,
		hub:         hub,
		connection:  connection,
		messageType: messageType,
	}
}

var _ IStream = (*stream)(nil)

func (s *stream) ConnectionID() ConnectionID {
	return s.connection.GetID()
}

func (s *stream) Connection() IConnection {
	return s.connection
}

func (s *stream) MessageType() int {
	return s.messageType
}

func (s *stream) Context() context.Context {
	return s.ctx
}

func (s *stream) Send(message []byte) error {
	return s.hub.Send(s.ctx, s.messageType, message, s.connection.GetID())
}

// SendProtoMessage 将proto.Message转化为二进制、JSON之后发送
func (s *stream) SendProtoMessage(message proto.Message) error {
	return s.hub.SendProtoMessage(s.ctx, s.messageType, message, s.connection.GetID())
}
