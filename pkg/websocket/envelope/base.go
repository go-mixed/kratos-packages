package envelope

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
)

type envelopeMarshaler interface {
	Marshal(e base.IEnvelope) ([]byte, error)
	Unmarshal(data []byte) (base.IEnvelope, error)
}

var envelopeCoding map[string]envelopeMarshaler = map[string]envelopeMarshaler{}

// RegisterEnvelopeEncoding 注册envelope的Marshaler和Unmarshaler
func RegisterEnvelopeEncoding(module string, encoding envelopeMarshaler) {
	envelopeCoding[module] = encoding
}

type EnvelopeBase struct {
	ID          string `json:"id" msgpack:"id"`
	AppID       string `json:"app_id" msgpack:"app_id"`
	Message     []byte `json:"message" msgpack:"message"`
	MessageType int    `json:"type" msgpack:"type"`
	Attempts    int    `json:"attempts" msgpack:"attempts"`

	// 由于消息可能来源于其他节点，为了保证日志的完整性，context需要记录一些trace信息
	Context string `json:"context" msgpack:"context"`
}

func (e *EnvelopeBase) GetMessageType() int {
	return e.MessageType
}

func (e *EnvelopeBase) GetMessage() []byte {
	return e.Message
}

func (e *EnvelopeBase) GetID() string {
	return e.ID
}

func (e *EnvelopeBase) GetAttempts() int {
	return e.Attempts
}

func (e *EnvelopeBase) GetOriginalAppID() string {
	return e.AppID
}

func (e *EnvelopeBase) GetContext(parentCtx context.Context) context.Context {
	return utils.DeserializeContext(parentCtx, e.Context)
}

func (e *EnvelopeBase) SetMessageType(messageType int) {
	e.MessageType = messageType
}

func (e *EnvelopeBase) SetMessage(message []byte) {
	e.Message = message
}

func (e *EnvelopeBase) SetID(id string) {
	e.ID = id
}

func (e *EnvelopeBase) SetOriginalAppID(appID string) {
	e.AppID = appID
}

func (e *EnvelopeBase) SetAttempts(attempts int) {
	e.Attempts = attempts
}

func (e *EnvelopeBase) SetContext(context context.Context) {
	e.Context = utils.SerializeContext(context)
}

type EnvelopeBaseBuilder[B any, E base.IEnvelope] struct {
	builder  B
	envelope E
}

func NewEnvelopeBaseBuilder[B any, E base.IEnvelope](inheritedBuilder B, inheritedEnvelope E) *EnvelopeBaseBuilder[B, E] {
	return &EnvelopeBaseBuilder[B, E]{
		envelope: inheritedEnvelope,
		builder:  inheritedBuilder,
	}
}

func (b *EnvelopeBaseBuilder[B, E]) WithID(id string) B {
	b.envelope.SetID(id)
	return b.builder
}

func (b *EnvelopeBaseBuilder[B, E]) WithAppID(appID string) B {
	b.envelope.SetOriginalAppID(appID)
	return b.builder
}

func (b *EnvelopeBaseBuilder[B, E]) WithAttempts(attempts int) B {
	b.envelope.SetAttempts(attempts)
	return b.builder
}

func (b *EnvelopeBaseBuilder[B, E]) WithMessageType(messageType int) B {
	b.envelope.SetMessageType(messageType)
	return b.builder
}

func (b *EnvelopeBaseBuilder[B, E]) WithMessage(message []byte) B {
	b.envelope.SetMessage(message)
	return b.builder
}

func (b *EnvelopeBaseBuilder[B, E]) WithContext(context context.Context) B {
	b.envelope.SetContext(context)
	return b.builder
}

func (b *EnvelopeBaseBuilder[B, E]) Build() E {
	if b.envelope.GetID() == "" {
		b.envelope.SetID(uuid.New().String())
	}
	if b.envelope.GetMessageType() == 0 {
		b.envelope.SetMessageType(base.TextMessage)
	}
	return b.envelope
}

type envelopeWrapper struct {
	Envelope  json.RawMessage `json:"envelope"`
	ClassName string          `json:"class_name"`
}

// MarshalIEnvelope 将IEnvelope转换为json
func MarshalIEnvelope(e base.IEnvelope) ([]byte, error) {
	var module string
	var bob json.RawMessage

	className := utils.GetClassName(e)
	if encoding, ok := envelopeCoding[className]; ok {
		bob, _ = encoding.Marshal(e)
		module = className
	} else {
		return nil, errors.New("marshal invalid envelope")
	}

	return json.Marshal(&envelopeWrapper{
		Envelope:  bob,
		ClassName: module,
	})
}

// UnmarshalIEnvelope 将json转换为IEnvelope
func UnmarshalIEnvelope(data []byte) (base.IEnvelope, error) {
	var w envelopeWrapper
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}

	if encoding, ok := envelopeCoding[w.ClassName]; ok {
		return encoding.Unmarshal(w.Envelope)
	}

	return nil, errors.New("unmarshal invalid envelope")
}
