package auth

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

type authMiddlewareFunc func(ctx context.Context, transporter transport.Transporter, requestToken string) (auth.IAuth, error)

// NewAuthMiddleware 用于Kratos http server的auth中间件
func NewAuthMiddleware(authFunc authMiddlewareFunc, logger log.Logger) middleware.Middleware {
	logHelper := log.NewModuleHelper(logger, "middleware/http")
	return func(nextHandler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			l := logHelper.WithContext(ctx)
			transporter, ok := transport.FromServerContext(ctx)
			if !ok {
				l.Error("wrong transport context for auth middleware")
				return nil, auth.ErrWrongContext
			}

			authHeaderValue := strings.TrimSpace(transporter.RequestHeader().Get(auth.AuthorizationHeader))
			if authHeaderValue == "" {
				l.Errorf("requestToken is missing of \"%s\"", transporter.Operation())
				return nil, auth.ErrMissingToken
			}

			// 从请求头中获取token，有Bearer开头的话去掉
			var requestToken string
			if !strings.HasPrefix(authHeaderValue, auth.BearerWord) {
				requestToken = authHeaderValue
			} else {
				requestToken = strings.TrimSpace(authHeaderValue[len(auth.BearerWord):])
			}

			authImpl, err := authFunc(ctx, transporter, requestToken)
			if err != nil {
				l.Errorf("requestToken authFunc of \"%s\" err: %v", transporter.Operation(), err)
				return nil, err
			}

			ctx = auth.NewContext(ctx, authImpl)
			return nextHandler(ctx, req)
		}
	}
}
