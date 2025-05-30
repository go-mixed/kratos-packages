package envelope

import (
	"context"
	"github.com/gorilla/websocket"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
)

type ObsoleteEnvelope struct {
	*EnvelopeBase
	ObsoleteConnectionID base.ConnectionID `json:"obsolete_connection_id" msgpack:"obsolete_connection_id"`
}

func init() {
	// 将ObsoleteEnvelope注册到websocket.EnvelopeEncoding中
	RegisterEnvelopeEncoding(utils.GetClassName((*ObsoleteEnvelope)(nil)), &ObsoleteEnvelope{})
}

var _ base.IEnvelope = (*ObsoleteEnvelope)(nil)

func NewObsoleteEnvelope(ctx context.Context, connectionId base.ConnectionID) *ObsoleteEnvelope {
	return NewObsoleteEnvelopeBuilder().
		WithMessageType(websocket.CloseMessage).
		WithMessage([]byte("obsolete")).
		WithContext(ctx).
		WithObsoleteConnectionID(connectionId).
		Build()
}

func (e *ObsoleteEnvelope) GetSendingConnections(conns base.IConnections) base.IConnections {
	return conns.MGet(e.ObsoleteConnectionID)
}

func (e *ObsoleteEnvelope) Copy() base.IEnvelope {
	_e := *e.EnvelopeBase
	return &ObsoleteEnvelope{
		EnvelopeBase:         &_e,
		ObsoleteConnectionID: e.ObsoleteConnectionID,
	}
}

func (e *ObsoleteEnvelope) Marshal(envelope base.IEnvelope) ([]byte, error) {
	return marshal[*ObsoleteEnvelope](envelope)
}

func (e *ObsoleteEnvelope) Unmarshal(data []byte) (base.IEnvelope, error) {
	return unmarshal[*ObsoleteEnvelope](data)
}

type obsoleteEnvelopeBuilder struct {
	obsoleteEnvelope *ObsoleteEnvelope
	*EnvelopeBaseBuilder[*obsoleteEnvelopeBuilder, *ObsoleteEnvelope]
}

func NewObsoleteEnvelopeBuilder() *obsoleteEnvelopeBuilder {
	b := &obsoleteEnvelopeBuilder{
		obsoleteEnvelope: &ObsoleteEnvelope{
			EnvelopeBase: &EnvelopeBase{},
		},
	}
	b.EnvelopeBaseBuilder = NewEnvelopeBaseBuilder(b, b.obsoleteEnvelope)
	return b
}

func (b *obsoleteEnvelopeBuilder) WithObsoleteConnectionID(connectionID base.ConnectionID) *obsoleteEnvelopeBuilder {
	b.obsoleteEnvelope.ObsoleteConnectionID = connectionID
	return b
}
