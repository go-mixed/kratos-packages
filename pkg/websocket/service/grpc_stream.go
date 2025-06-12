package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	wsProto "gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/proto"
	"strings"
)

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

func MakeGrpcResponse(messageId, serviceName, methodName string, responseData proto.Message, err error) *wsProto.WebsocketGrpcResponse {
	var data *anypb.Any
	if responseData != nil {
		data, _ = anypb.New(responseData)
	}
	return &wsProto.WebsocketGrpcResponse{
		Service:   serviceName,
		Method:    methodName,
		Type:      strings.ReplaceAll(data.GetTypeUrl(), "type.googleapis.com/", ""),
		MessageId: utils.Ptr(messageId),

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

type grpcStream struct {
	ctx         context.Context
	grpcHandler *GrpcService
	connection  base.IConnection
	request     *wsProto.WebsocketGrpcRequest
	method      grpcMethodInfo
	messageType int
}

var _ grpc.ServerStream = (*grpcStream)(nil)

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
	message, ok := m.(proto.Message)
	if !ok {
		return errors.InternalServer("INTERNAL", "invalid response type")
	}
	return g.grpcHandler.sendResponse(g.ctx, g.connection, g.messageType, g.request, message)
}

func (g *grpcStream) RecvMsg(m any) error {
	return getGrpcRequestData(g.request)(m)
}
