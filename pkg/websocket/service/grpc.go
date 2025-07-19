package service

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	wsProto "gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/proto"
)

type GrpcService struct {
	app      *app.App
	logger   *log.Helper
	services map[string]*grpcServiceInfo

	hub base.IHub
}

type grpcMethodInfo struct {
	serviceName string
	server      any

	methodDesc     *grpc.MethodDesc
	streamDesc     *grpc.StreamDesc
	requestStream  bool
	responseStream bool
}

// serviceInfo wraps information about a service. It is very similar to
// ServiceDesc and is constructed from it for internal purposes.
type grpcServiceInfo struct {
	// Contains the implementation for the methods in this service.
	serviceImpl any
	methods     map[string]*grpc.MethodDesc
	streams     map[string]*grpc.StreamDesc
	mdata       any
}

func NewGrpcService(
	app *app.App,
	logger log.Logger,
	hub base.IHub,
) *GrpcService {
	return &GrpcService{
		app:      app,
		logger:   log.NewModuleHelper(logger, "websocket/service/grpc"),
		hub:      hub,
		services: make(map[string]*grpcServiceInfo),
	}
}

var _ base.IHandler = (*GrpcService)(nil)

func (g *GrpcService) RegisterService(desc *grpc.ServiceDesc, impl any) {
	info := &grpcServiceInfo{
		serviceImpl: impl,
		methods:     make(map[string]*grpc.MethodDesc),
		streams:     make(map[string]*grpc.StreamDesc),
		mdata:       desc.Metadata,
	}
	for i := range desc.Methods {
		d := &desc.Methods[i]
		info.methods[d.MethodName] = d
	}
	for i := range desc.Streams {
		d := &desc.Streams[i]
		info.streams[d.StreamName] = d
	}
	g.services[desc.ServiceName] = info
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
	request := &wsProto.WebsocketRequest{}

	if err := base.ProtoUnmarshal(messageType, message, request); err != nil {
		g.sendError(ctx, connection, messageType, request, err)
		return nil
	}

	ctx = wsProto.NewContext(ctx, request)

	method, err := g.getServiceMethod(request.GetService(), request.GetMethod())
	if err != nil {
		g.sendError(ctx, connection, messageType, request, err)
		return nil
	}

	// 在作用域中阻塞调用
	func(ctx context.Context, connection base.IConnection, messageType int, request *wsProto.WebsocketRequest, method grpcMethodInfo) {
		// ctx的生命周期只在响应内有效
		sessionCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		var response proto.Message
		var err error
		_grpcStream := &grpcStream{ctx: sessionCtx, grpcHandler: g, connection: connection, request: request, method: method, messageType: messageType}

		if method.requestStream || method.responseStream {
			err = g.callStreamService(_grpcStream)
		} else {
			response, err = g.callService(_grpcStream)
		}

		if err != nil {
			g.sendError(sessionCtx, connection, messageType, request, err)
		} else if !utils.IsNil(response) {
			// 此处必须使用IsNil来判断，因为即使在RPC函数中返回nil，但是会经过几次interface封装（callService中），会导致response != nil
			// 当response为空时，不需要发送响应。如果希望异步回复，返回nil
			if err = g.sendResponse(sessionCtx, connection, messageType, request, response); err != nil {
				g.sendError(sessionCtx, connection, messageType, request, err)
			}
		}

	}(ctx, connection, messageType, request, method)

	return nil
}

func (g *GrpcService) OnSendMessage(ctx context.Context, envelope base.IEnvelope, connections base.SentConnections) error {
	return nil
}

func (g *GrpcService) OnStart(ctx context.Context) {

}

func (g *GrpcService) OnStop(ctx context.Context) {

}

func (g *GrpcService) getServiceMethod(serviceName string, methodName string) (grpcMethodInfo, error) {
	var pbService grpcMethodInfo
	service, ok := g.services[serviceName]
	if !ok {
		return pbService, errors.NotFound("", fmt.Sprintf("service \"%s\" not found", serviceName))
	}

	method, ok1 := service.methods[methodName]
	stream, ok2 := service.streams[methodName]

	if !ok1 && !ok2 {
		return pbService, errors.NotFound("", fmt.Sprintf("service \"%s\" method \"%s\" not found", serviceName, methodName))
	}

	pbService.requestStream = lo.IfF(stream != nil, func() bool {
		return stream.ClientStreams
	}).Else(false)
	pbService.responseStream = lo.IfF(stream != nil, func() bool {
		return stream.ServerStreams
	}).Else(false)
	pbService.serviceName = serviceName
	pbService.server = service.serviceImpl
	pbService.methodDesc = method
	pbService.streamDesc = stream

	return pbService, nil
}

// callService 同步调用服务
func (g *GrpcService) callService(stream *grpcStream) (proto.Message, error) {
	data := getGrpcRequestData(stream.request)
	response, err := stream.method.methodDesc.Handler(stream.method.server, stream.ctx, data, nil)
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
func (g *GrpcService) callStreamService(stream *grpcStream) error {
	return stream.method.streamDesc.Handler(stream.method.server, stream)
}

func (g *GrpcService) sendResponse(ctx context.Context, connection base.IConnection, messageType int, request *wsProto.WebsocketRequest, responseData proto.Message) error {
	// 发送响应
	response := wsProto.MakeWebsocketResponse(request.GetMessageId(), request.Service, request.Method, responseData, nil)
	return g.hub.SendProtoMessage(ctx, messageType, response, connection.GetID())
}

func (g *GrpcService) sendError(ctx context.Context, connection base.IConnection, messageType int, request *wsProto.WebsocketRequest, err error) {
	response := wsProto.MakeWebsocketResponse(request.GetMessageId(), request.Service, request.Method, nil, err)

	err = g.hub.SendProtoMessage(ctx, messageType, response, connection.GetID())
	if err != nil {
		g.logger.WithContext(ctx).Errorf("send error response error: %v", err)
	}
}
