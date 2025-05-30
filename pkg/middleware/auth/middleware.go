package auth

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"

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

			requestToken := auth.StripAuthorization(transporter.RequestHeader().Get(auth.AuthorizationHeader))
			if requestToken == "" {
				l.Errorf("cannot get authorization")

				return nil, auth.ErrMissingToken
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
