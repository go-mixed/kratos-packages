package service

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/worker"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	wsProto "gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/proto"
)

type GrpcService struct {
	app      *app.App
	worker   worker.IWorker
	logger   *log.Helper
	services map[string]*grpc.ServiceDesc

	hub base.IHub
}

type pbService struct {
	serviceName string
	server      any
	methodDesc  grpc.MethodDesc
	streamDesc  grpc.StreamDesc
	stream      bool
}

func NewGrpcService(
	app *app.App,
	logger log.Logger,
	worker worker.IWorker,
	hub base.IHub,
) *GrpcService {
	return &GrpcService{
		app:      app,
		logger:   log.NewModuleHelper(logger, "websocket/service/grpc"),
		worker:   worker,
		hub:      hub,
		services: make(map[string]*grpc.ServiceDesc),
	}
}

var _ base.IHandler = (*GrpcService)(nil)

func (g *GrpcService) RegisterService(desc *grpc.ServiceDesc, impl any) {
	var _desc = new(grpc.ServiceDesc)
	// copy
	*_desc = *desc
	_desc.HandlerType = impl

	g.services[_desc.ServiceName] = _desc
}

func (g *GrpcService) OnPong(ctx context.Context, connection base.IConnection) error {
	return nil
}

func (g *GrpcService) OnError(ctx context.Context, connection base.IConnection, err error) {

}

func (g *GrpcService) OnClose(ctx context.Context, connection base.IConnection, i int, s string) error {
	return nil
}

func (g *GrpcService) OnConnect(ctx context.Context, connection base.IConnection) error {
	return nil
}

func (g *GrpcService) OnDisconnect(ctx context.Context, connection base.IConnection) error {
	return nil
}

func (g *GrpcService) OnRecvMessage(ctx context.Context, connection base.IConnection, messageType int, message []byte) error {
	request := &wsProto.WebsocketGrpcRequest{}

	if err := g.unmarshal(messageType, message, request); err != nil {
		g.sendError(ctx, connection, messageType, request, base.ErrInvalidEnvelope)
		return nil
	}

	pbService, err := g.getServiceMethod(request.GetService(), request.GetMethod())
	if err != nil {
		g.sendError(ctx, connection, messageType, request, err)
		return nil
	}

	if pbService.stream {

	} else {
		// 同步调用
		if response, err := g.callService(ctx, request, pbService); err != nil {
			g.sendError(ctx, connection, messageType, request, err)
		} else if response != nil {
			if err = g.sendResponse(ctx, connection, messageType, request, response); err != nil {
				g.sendError(ctx, connection, messageType, request, err)
			}
		}
	}

	return nil
}

func (g *GrpcService) OnSendMessage(ctx context.Context, envelope base.IEnvelope, connections base.SentConnections) error {
	return nil
}

func (g *GrpcService) OnStart(ctx context.Context) {

}

func (g *GrpcService) OnStop(ctx context.Context) {

}

func (g *GrpcService) unmarshal(messageType int, data []byte, in proto.Message) error {
	if messageType == base.BinaryMessage {
		return proto.Unmarshal(data, in)
	} else if messageType == base.TextMessage {
		return json.UnmarshalOptions.Unmarshal(data, in)
	}
	return nil
}

func (g *GrpcService) marshal(messageType int, message proto.Message) []byte {
	var bytes []byte
	if messageType == base.BinaryMessage {
		bytes, _ = proto.Marshal(message)
	} else if messageType == base.TextMessage {
		bytes, _ = json.MarshalOptions.Marshal(message)
	}
	return bytes
}

func (g *GrpcService) getServiceMethod(serviceName string, methodName string) (pbService, error) {
	var pbService pbService
	service, ok := g.services[serviceName]
	if !ok {
		return pbService, errors.NotFound("", fmt.Sprintf("service \"%s\" not found", serviceName))
	}

	method, ok1 := lo.Find(service.Methods, func(method grpc.MethodDesc) bool {
		return method.MethodName == methodName
	})

	stream, ok2 := lo.Find(service.Streams, func(stream grpc.StreamDesc) bool {
		return stream.StreamName == methodName
	})

	if !ok1 && !ok2 {
		return pbService, errors.NotFound("METHOD", fmt.Sprintf("service \"%s\" method \"%s\" not found", serviceName, methodName))
	} else if ok2 {
		pbService.stream = true
	}

	pbService.serviceName = service.ServiceName
	pbService.server = service.HandlerType
	pbService.methodDesc = method
	pbService.streamDesc = stream

	return pbService, nil
}

// getRequestData 获取请求的Data，转成proto.Message
func (g *GrpcService) getRequestData(request *wsProto.WebsocketGrpcRequest) func(any) error {
	return func(in any) error {
		msg, ok := in.(proto.Message)
		if !ok {
			return errors.InternalServer("INTERNAL", "invalid data type")
		}
		return request.GetData().UnmarshalTo(msg)
	}
}

// callService 同步调用服务
func (g *GrpcService) callService(ctx context.Context, request *wsProto.WebsocketGrpcRequest, pbService pbService) (proto.Message, error) {
	response, err := pbService.methodDesc.Handler(pbService.server, ctx, g.getRequestData(request), nil)
	if err != nil {
		return nil, err
	}
	_response, ok := response.(proto.Message)
	if !ok {
		return nil, errors.InternalServer("INTERNAL", "invalid response type")
	}
	return _response, nil
}

// callStreamService 调用流式服务
func (g *GrpcService) callStreamService(ctx context.Context, request *wsProto.WebsocketGrpcRequest, pbService pbService) error {
	//g.worker.WithContext(ctx).Submit(func(ctx context.Context) {
	return nil
	//})
}

func (g *GrpcService) sendError(ctx context.Context, connection base.IConnection, messageType int, request *wsProto.WebsocketGrpcRequest, err error) {
	response := &wsProto.WebsocketGrpcResponse{
		MessageId: lo.If(request.MessageId != nil && request.GetMessageId() != "", request.MessageId).ElseF(func() *string {
			return utils.Ptr(uuid.New().String())
		}),
		Code: lo.IfF(err != nil, func() int32 {
			res := int32(errors.Code(err))
			if res == 0 {
				res = 400
			}
			return res
		}).Else(0),
		Message: lo.IfF(err != nil, func() string {
			return err.Error()
		}).Else(""),
	}

	bytes := g.marshal(messageType, response)

	if messageType == base.BinaryMessage {
		err = g.hub.SendBinary(ctx, bytes, connection.GetID())
	} else if messageType == base.TextMessage {
		err = g.hub.SendText(ctx, string(bytes), connection.GetID())
	}

	if err != nil {
		g.logger.WithContext(ctx).Errorf("send error response error: %v", err)
	}
}

func (g *GrpcService) sendResponse(ctx context.Context, connection base.IConnection, messageType int, request *wsProto.WebsocketGrpcRequest, responseData proto.Message) error {
	data, _ := anypb.New(responseData)

	response := &wsProto.WebsocketGrpcResponse{
		MessageId: lo.If(request.MessageId != nil && request.GetMessageId() != "", request.MessageId).ElseF(func() *string {
			return utils.Ptr(uuid.New().String())
		}),
		Code:    0,
		Message: "",
		Data:    data,
	}

	bytes := g.marshal(messageType, response)
	var err error
	if messageType == base.BinaryMessage {
		err = g.hub.SendBinary(ctx, bytes, connection.GetID())
	} else if messageType == base.TextMessage {
		err = g.hub.SendText(ctx, string(bytes), connection.GetID())
	}

	return err
}
