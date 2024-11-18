package requestid

import (
	"context"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
)

// Server 实例化trace中间件，先尝试从，会添加trace
func Server() middleware.Middleware {
	return func(nextHandler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			var reqId string
			// 优先从x-md中获取requestId
			if _metadata, ok := metadata.FromServerContext(ctx); ok {
				reqId = _metadata.Get(requestid.HeaderXRequestID)
			}
			header, ok := transport.FromServerContext(ctx)
			// 尝试再从header中获取requestId
			if ok && reqId == "" {
				reqId = header.RequestHeader().Get(requestid.HeaderXRequestID)
			}
			// 最后尝试从context中获取requestId
			if reqId == "" {
				reqId = requestid.FromContext(ctx)
			}
			// 如果3种方法都没取到，就生成requestId
			if reqId == "" {
				reqId = requestid.GenerateRequestId()
			}

			// 将requestId添加到响应头中
			if ok {
				header.ReplyHeader().Set(requestid.HeaderXRequestID, reqId)
			}

			// 将requestId添加到context中
			ctx = requestid.NewContext(ctx, reqId)
			reply, err = nextHandler(ctx, req)

			if err == nil {
				fillReqIdAppId(ctx, reply, reqId)
			}

			return
		}
	}
}
