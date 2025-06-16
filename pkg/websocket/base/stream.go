package base

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	wsProto "gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/proto"
	"strings"
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

// MakeWebsocketResponse 创建websocket响应
func MakeWebsocketResponse(messageId, serviceName, methodName string, responseData proto.Message, err error) *wsProto.WebsocketResponse {
	var data *anypb.Any
	if a, ok := responseData.(*anypb.Any); ok {
		data = a
	} else if responseData != nil {
		data, _ = anypb.New(responseData)
	}
	return &wsProto.WebsocketResponse{
		MessageId: lo.IfF(messageId != "", func() *string {
			return &messageId
		}).Else(nil),
		Service: serviceName,
		Method:  methodName,
		Type: lo.IfF(data != nil, func() string {
			return strings.ReplaceAll(data.GetTypeUrl(), "type.googleapis.com/", "")
		}).Else(""),

		Code: lo.IfF(err != nil, func() int32 {
			if code := errors.Code(err); code > 0 {
				return int32(code)
			}
			return 400
		}).Else(0),
		Message: lo.IfF(err != nil, func() string {
			return err.Error()
		}).Else(""),
		Data: data,
	}
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
