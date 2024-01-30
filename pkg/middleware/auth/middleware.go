package auth

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

type authMiddlewareFunc func(ctx context.Context, transporter transport.Transporter, token string) (auth.IAuth, error)

func NewAuthMiddleware(authFunc authMiddlewareFunc, logger log.Logger) middleware.Middleware {
	logHelper := log.NewModuleHelper(logger, "middleware/http")
	return func(nextHandler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			l := logHelper.WithContext(ctx)
			transporter, ok := transport.FromServerContext(ctx)
			if !ok {
				l.Error("wrong transport context for auth middleware")
				return nil, auth.ErrWrongContext
			}
			auths := strings.SplitN(transporter.RequestHeader().Get(auth.AuthorizationKey), " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], auth.BearerWord) {
				l.Errorf("token is missing of \"%s\"", transporter.Operation())
				return nil, auth.ErrMissingToken
			}
			token := auths[1]

			authImpl, err := authFunc(ctx, transporter, token)
			if err != nil {
				l.Errorf("token authFunc of \"%s\" err: %v", transporter.Operation(), err)
				return nil, err
			}

			ctx = auth.NewContext(ctx, authImpl)
			return nextHandler(ctx, req)
		}
	}
}
