package auth

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
)

type authKey struct{}

// NewContext put auth info into context
func NewContext(ctx context.Context, info IAuth) context.Context {
	return context.WithValue(ctx, authKey{}, info)
}

func FromContext(ctx context.Context) (IAuth, bool) {
	info, ok := ctx.Value(authKey{}).(IAuth)
	return info, ok
}

func GetAndValidate(ctx context.Context, guardName string) (IAuth, error) {
	session, ok := FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("", "请先登录相关账号")
	} else if session.GetGuardName() != guardName {
		return nil, errors.Forbidden(guardName, "当前登录的账号没有权限操作这个资源")
	}

	return session, nil
}
