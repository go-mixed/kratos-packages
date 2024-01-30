package websocket

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"net/http"
)

type IHub interface {
	Closed() bool
	Close(exitMessage IEnvelope)
	Start(context.Context)

	Authenticate(r *http.Request) (auth.IAccessToken, auth.IGuard, error)
	Register(session *Session) error
	Unregister(session *Session) error
	SetServer(o *Server)

	Send(ctx context.Context, envelope IEnvelope) error
	SendGuard(ctx context.Context, guard auth.IGuard, message []byte, services ...string) error
	Broadcast(ctx context.Context, matches map[string]any, message []byte) error
}

type emptyHub struct{}

var _ IHub = (*emptyHub)(nil)

func newEmptyHub() *emptyHub {
	return &emptyHub{}
}

func (e emptyHub) SetServer(o *Server) {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Closed() bool {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Authenticate(r *http.Request) (auth.IAccessToken, auth.IGuard, error) {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Register(session *Session) error {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Unregister(session *Session) error {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Start(context.Context) {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Close(exitMessage IEnvelope) {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Send(ctx context.Context, envelope IEnvelope) error {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) SendGuard(ctx context.Context, guard auth.IGuard, message []byte, services ...string) error {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}

func (e emptyHub) Broadcast(ctx context.Context, matches map[string]any, message []byte) error {
	panic("MUST be with \"WithHub\" when creating websocket.Server.")
}
