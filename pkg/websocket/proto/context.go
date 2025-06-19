package proto

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"strings"
)

type websocketRequestKey struct{}

// MakeWebsocketResponse 创建websocket响应
func MakeWebsocketResponse(messageId, serviceName, methodName string, responseData proto.Message, err error) *WebsocketResponse {
	var data *anypb.Any
	if a, ok := responseData.(*anypb.Any); ok {
		data = a
	} else if responseData != nil {
		data, _ = anypb.New(responseData)
	}
	return &WebsocketResponse{
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

func FromContext(ctx context.Context) (*WebsocketRequest, bool) {
	val, ok := ctx.Value(websocketRequestKey{}).(*WebsocketRequest)
	return val, ok && val != nil
}

func NewContext(ctx context.Context, r *WebsocketRequest) context.Context {
	ctx = context.WithValue(ctx, websocketRequestKey{}, r)
	return ctx
}
