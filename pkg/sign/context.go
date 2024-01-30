package sign

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
)

type signKey struct{}

// NewContext put sign info into context
func NewContext(ctx context.Context, info auth.IThirdParty) context.Context {
	return context.WithValue(ctx, signKey{}, info)
}

func FromContext(ctx context.Context) (auth.IThirdParty, bool) {
	info, ok := ctx.Value(signKey{}).(auth.IThirdParty)
	return info, ok
}

func MustFromContext(ctx context.Context) auth.IThirdParty {
	info, ok := FromContext(ctx)
	if !ok {
		return auth.NewThirdParty(0, "", "")
	}
	return info
}
