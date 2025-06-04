package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	wsProto "gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/proto"
	"strings"
)

func grpcUnmarshal(messageType int, data []byte, in proto.Message) error {
	if messageType == base.BinaryMessage {
		return proto.Unmarshal(data, in)
	} else if messageType == base.TextMessage {
		return json.UnmarshalOptions.Unmarshal(data, in)
	}
	return nil
}

func grpcMarshal(messageType int, message proto.Message) []byte {
	var bytes []byte
	if messageType == base.BinaryMessage {
		bytes, _ = proto.Marshal(message)
	} else if messageType == base.TextMessage {
		bytes, _ = json.MarshalOptions.Marshal(message)
	}
	return bytes
}

// getGrpcRequestData 获取grpc请求的Data，转成proto.Message
func getGrpcRequestData(request *wsProto.WebsocketGrpcRequest) func(any) error {
	return func(in any) error {
		msg, ok := in.(proto.Message)
		if request.Data == nil || !ok {
			return nil
		}

		if request.Data.TypeUrl == "" {
			request.Data.TypeUrl = "type.googleapis.com/" + string(msg.ProtoReflect().Descriptor().FullName())
		}

		if err := request.GetData().UnmarshalTo(msg); err != nil {
			return err
		}

		utils.PtrElement(in).Set(utils.PtrElement(msg))
		return nil
	}
}

func sendGrpcResponse(ctx context.Context, hub base.IHub, connection base.IConnection, messageType int, request *wsProto.WebsocketGrpcRequest, responseData proto.Message) error {
	data, _ := anypb.New(responseData)

	url := data.GetTypeUrl()
	url = strings.ReplaceAll(url, "type.googleapis.com/", "")

	response := &wsProto.WebsocketGrpcResponse{
		Service: request.Service,
		Method:  request.Method,
		Type:    url,
		MessageId: lo.If(request.MessageId != nil && request.GetMessageId() != "", request.MessageId).ElseF(func() *string {
			return utils.Ptr(uuid.New().String())
		}),
		Code:    0,
		Message: "",
		Data:    data,
	}

	bytes := grpcMarshal(messageType, response)
	var err error
	if messageType == base.BinaryMessage {
		err = hub.SendBinary(ctx, bytes, connection.GetID())
	} else if messageType == base.TextMessage {
		err = hub.SendText(ctx, string(bytes), connection.GetID())
	}

	return err
}

type grpcStream struct {
	ctx         context.Context
	hub         base.IHub
	connection  base.IConnection
	request     *wsProto.WebsocketGrpcRequest
	method      grpcMethodInfo
	messageType int
}

func (g *grpcStream) SetHeader(md metadata.MD) error {
	return nil
}

func (g *grpcStream) SendHeader(md metadata.MD) error {
	return nil
}

func (g *grpcStream) SetTrailer(md metadata.MD) {

}

func (g *grpcStream) Context() context.Context {
	return g.ctx
}

func (g *grpcStream) SendMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return errors.InternalServer("INTERNAL", "invalid response type")
	}

	return sendGrpcResponse(g.ctx, g.hub, g.connection, g.messageType, g.request, msg)
}

func (g *grpcStream) RecvMsg(m any) error {
	return getGrpcRequestData(g.request)(m)
}
