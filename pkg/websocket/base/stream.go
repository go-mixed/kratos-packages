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

func (s *stream) MessageType() int {
	return s.messageType
}

func (s *stream) Context() context.Context {
	return s.ctx
}

func (s *stream) Send(message []byte) error {
	var err error
	if s.messageType == BinaryMessage {
		err = s.hub.SendBinary(s.ctx, message, s.connection.GetID())
	} else if s.messageType == TextMessage {
		err = s.hub.SendText(s.ctx, string(message), s.connection.GetID())
	}
	return err
}

// SendProtoMessage 发送原始proto消息，不会经过封装
func (s *stream) SendProtoMessage(message proto.Message) error {
	bytes := ProtoMarshal(s.messageType, message)
	return s.Send(bytes)
}
