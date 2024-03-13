package rbac

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

type RbacMiddlewareFunc func(ctx context.Context, guard auth.IGuard, args ...string) (bool, error)

func NewRbacMiddleware(rbacFunc RbacMiddlewareFunc, logger log.Logger) func(ctx context.Context, req any, arguments ...string) (any, error) {
	logHelper := log.NewModuleHelper(logger, "middleware/rbac")
	return func(ctx context.Context, req interface{}, arguments ...string) (any, error) {
		l := logHelper.WithContext(ctx)
		var guard auth.IGuard
		if user, _ := auth.FromContext(ctx); user == nil {
			guard = &auth.Guard{
				GuardName:       "anonymous",
				AuthorizationID: 0,
			}
		} else {
			guard = user.GetGuardModel()
		}

		allowed, err := rbacFunc(ctx, guard, arguments...)
		if err != nil {
			l.Errorf("rbacFunc fail, guard: %s:%d args: %v err: %v", guard.GetGuardName(), guard.GetAuthorizationID(), arguments, err)
			return nil, err
		}

		if !allowed {
			l.Debugf("rbacFunc deny, guard: %s:%d args: %v", guard.GetGuardName(), guard.GetAuthorizationID(), arguments)
			return nil, auth.ErrForbidden
		}

		return req, nil
	}
}
