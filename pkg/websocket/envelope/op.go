package envelope

import (
	"github.com/google/uuid"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
)

type OperationEnvelope struct {
	*Envelope
}

var _ base.IEnvelope = (*OperationEnvelope)(nil)

// NewOperationEnvelope 目前仅仅用于Close、Ping这2个原生的消息。
func NewOperationEnvelope(messageType int, message []byte) *OperationEnvelope {
	if messageType != base.CloseMessage && messageType != base.PingMessage {
		panic("invalid messageType")
	}

	return &OperationEnvelope{
		Envelope: &Envelope{
			EnvelopeBase: &EnvelopeBase{
				ID:          uuid.New().String(),
				MessageType: messageType,
				Message:     message,
				Attempts:    0,
			},
		},
	}
}

func (o *OperationEnvelope) GetSendingConnections(conns base.IConnections) base.IConnections {
	return base.Connections{}
}

func (o *OperationEnvelope) Copy() base.IEnvelope {
	return &OperationEnvelope{}
}
