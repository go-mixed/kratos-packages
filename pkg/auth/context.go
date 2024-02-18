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
		return nil, errors.Unauthorized("", "Please get the access token first")
	} else if session.GetGuardModel().GetGuardName() != guardName {
		return nil, errors.Forbidden(guardName, "The guard name of access token is not matched")
	}

	return session, nil
}
