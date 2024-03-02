package auth

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"slices"
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

// GetAndValidate get auth info from context and validate the guard name
func GetAndValidate(ctx context.Context, guardNames ...string) (IAuth, error) {
	session, ok := FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("", "Please get the access token first")
	} else if !slices.Contains(guardNames, session.GetGuardModel().GetGuardName()) {
		return nil, errors.Forbidden("", "The guard name of access token is not matched")
	}

	return session, nil
}
