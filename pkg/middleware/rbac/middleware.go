package rbac

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

type rbacMiddlewareFunc func(ctx context.Context, guard auth.IGuard) (bool, error)

// NewRbacMiddleware 用于Kratos http server的rbac中间件
func NewRbacMiddleware(rbacFunc rbacMiddlewareFunc, logger log.Logger) middleware.Middleware {
	logHelper := log.NewModuleHelper(logger, "middleware/rbac")
	return func(nextHandler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (any, error) {
			l := logHelper.WithContext(ctx)
			var guard auth.IGuard
			if user, _ := auth.FromContext(ctx); user != nil {
				guard = user.GetGuardModel()
			}

			allowed, err := rbacFunc(ctx, guard)
			if err != nil {
				l.Errorf("rbacFunc fail, guard: %s:%d args: %v err: %v", guard.GetGuardName(), guard.GetAuthorizationID(), err)
				return nil, err
			}

			if !allowed {
				l.Debugf("rbacFunc deny, guard: %s:%d args: %v", guard.GetGuardName(), guard.GetAuthorizationID())
				return nil, auth.ErrForbidden
			}

			return nextHandler(ctx, req)
		}
	}
}
