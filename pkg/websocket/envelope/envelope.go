package envelope

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
)

// Envelope 广播/私人消息
type Envelope struct {
	*EnvelopeBase

	ConnectionIDs []base.ConnectionID `json:"connection_ids" msgpack:"connection_ids"`
	All           bool                `json:"all" msgpack:"all"`
}

func init() {
	// 将BroadcastEnvelope注册到websocket.EnvelopeEncoding中
	RegisterEnvelopeEncoding(utils.GetClassName((*Envelope)(nil)), &Envelope{})
}

var _ base.IEnvelope = (*Envelope)(nil)

func (e *Envelope) Marshal(envelope base.IEnvelope) ([]byte, error) {
	return marshal[*Envelope](envelope)
}

func (e *Envelope) Unmarshal(data []byte) (base.IEnvelope, error) {
	return unmarshal[*Envelope](data)
}

func (e *Envelope) GetSendingConnections(conns base.IConnections) base.IConnections {
	if e.All {
		return conns
	}
	return conns.MGet(e.ConnectionIDs...)
}

func (e *Envelope) Copy() base.IEnvelope {
	_e := *e.EnvelopeBase
	return &Envelope{
		EnvelopeBase: &_e,
	}
}

type envelopeBuilder struct {
	envelope *Envelope
	*EnvelopeBaseBuilder[*envelopeBuilder, *Envelope]
}

func NewEnvelopeBuilder() *envelopeBuilder {
	b := &envelopeBuilder{
		envelope: &Envelope{
			EnvelopeBase: &EnvelopeBase{},
		},
	}
	b.EnvelopeBaseBuilder = NewEnvelopeBaseBuilder(b, b.envelope)
	return b
}

// WithConnectionIDs 设置接收者，如果all为true则发送给所有连接，否则发送给指定的连接
func (b *envelopeBuilder) WithConnectionIDs(all bool, ids ...base.ConnectionID) *envelopeBuilder {
	b.envelope.All = all
	b.envelope.ConnectionIDs = ids
	return b
}
