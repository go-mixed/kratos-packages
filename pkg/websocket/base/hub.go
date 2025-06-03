package base

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"net/http"
)

type IHub interface {
	Running() bool
	GetConnections() IConnections
	GetAllConnectionIDs(ctx context.Context) []ConnectionID
	Send(ctx context.Context, envelope IEnvelope) error
	SendText(ctx context.Context, message string, connIds ...ConnectionID) error
	SendBinary(ctx context.Context, message []byte, connIds ...ConnectionID) error
	BroadcastText(ctx context.Context, message string) error
	BroadcastBinary(ctx context.Context, message []byte) error
	Close(ctx context.Context, exitMessage string, connIds ...ConnectionID)
}

type IAdapter interface {
	Authorize(ctx context.Context, r *http.Request) (auth.IAuth, error)
	InvokeConnection(ctx context.Context, session IConnection) (IConnection, error)
}
