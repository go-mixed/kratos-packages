package requestid

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
)

// Server 实例化trace中间件，先尝试从，会添加trace
func Server() middleware.Middleware {
	return func(nextHandler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			reqId := requestid.GenerateRequestId()
			if header, ok := transport.FromServerContext(ctx); ok {
				// 尝试先从header中获取requestId
				if _id := header.RequestHeader().Get(requestid.HeaderXRequestID); _id != "" {
					reqId = _id
				} else if _id = requestid.FromContext(ctx); _id != "" { // 再尝试从context中获取requestId
					reqId = _id
				}

				header.ReplyHeader().Set(requestid.HeaderXRequestID, reqId)
			}
			ctx = requestid.NewContext(ctx, reqId)
			reply, err = nextHandler(ctx, req)

			if err == nil {
				fillReqIdAppId(ctx, reply, reqId)
			}

			return
		}
	}
}
