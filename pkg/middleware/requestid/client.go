package requestid

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
)

func Client() middleware.Middleware {
	return func(nextHandler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			reqId := requestid.FromContext(ctx)
			if reqId == "" {
				reqId = requestid.GenerateRequestId()
			}

			header, ok := transport.FromClientContext(ctx)
			if ok {
				// 请求头中添加requestId
				header.RequestHeader().Set(requestid.HeaderXRequestID, reqId)
			}

			ctx = requestid.NewContext(ctx, reqId)
			reply, err = nextHandler(ctx, req)

			if ok {
				// 响应头中添加requestId
				if header.ReplyHeader().Get(requestid.HeaderXRequestID) == "" {
					header.ReplyHeader().Set(requestid.HeaderXRequestID, reqId)
				}
			}

			if err == nil {
				fillReqIdAppId(ctx, reply, reqId)
			}

			return
		}
	}
}
