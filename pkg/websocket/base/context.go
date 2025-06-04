package base

import (
	"context"
)

type connectionIdKey struct{}

func NewContext(ctx context.Context, connId ConnectionID) context.Context {
	return context.WithValue(ctx, connectionIdKey{}, connId)
}

func FromContext(ctx context.Context) (ConnectionID, bool) {
	conn, ok := ctx.Value(connectionIdKey{}).(ConnectionID)
	return conn, ok
}
