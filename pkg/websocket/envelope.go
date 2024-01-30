package websocket

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
)

type envelopeMarshaler interface {
	Marshal(e IEnvelope) ([]byte, error)
	Unmarshal(data []byte) (IEnvelope, error)
}

var envelopeCoding map[string]envelopeMarshaler = map[string]envelopeMarshaler{}

// RegisterEnvelopeEncoding 注册envelope的Marshaler和Unmarshaler
func RegisterEnvelopeEncoding(module string, encoding envelopeMarshaler) {
	envelopeCoding[module] = encoding
}

// IEnvelope
// 切勿在SetXXX后面返回IEnvelope，因为返回的是*Envelope，而不是其继承的子集，会导致其继承的struct中的数据全丢失。
// 所以Copy方法必须由继承者实现
type IEnvelope interface {
	// SetMessageType 设置envelope的消息类型，比如：text、binary、close等
	SetMessageType(messageType int)
	// SetMessage 设置envelope的消息内容
	SetMessage(message []byte)
	// SetID 设置envelope的ID，这个ID是唯一的，用于标识这个envelope
	SetID(id string)
	// SetOriginalAppID 设置原始的appID，表示这个envelope是从哪个appID发出的，在Subscribe的时候会排除
	SetOriginalAppID(appID string)
	// SetAttempts 设置envelope的尝试次数，
	//  注意：这个和WithAttempts的区别在于，这个是直接设置Attempts，而WithAttempts是复制一个新的envelope，然后设置Attempts
	SetAttempts(attempts int)
	// SetContext 设置envelope的context
	//  因为Envelope可能来源于其他节点，为了保证日志的链路跟踪，Envelope需要记录一些trace信息
	SetContext(context context.Context)

	// GetMessageType 获取envelope的消息类型，比如：text、binary、close等
	GetMessageType() int
	// GetMessage 获取envelope的消息内容
	GetMessage() []byte
	// GetID 获取envelope的ID，这个ID是唯一的，用于标识这个envelope
	GetID() string
	// GetOriginalAppID 获取原始的appID，表示这个envelope是从哪个appID发出的，在Subscribe的时候会排除
	GetOriginalAppID() string
	// GetAttempts 获取envelope的尝试次数
	GetAttempts() int
	// GetContext 获取envelope的context
	//  即使Envelope来源于其他节点，也可以获取到trace相关的信息，用于日志的链路跟踪
	GetContext(parentCtx context.Context) context.Context

	// GetExpectSessionIDs 获取期望的sessionID，如果返回nil，则表示不需要发送给任何session
	//  继承者需要根据自己的业务逻辑，返回期望的sessionID
	//  原始Envelope如果调用本方法，会panic
	GetExpectSessionIDs(sessions ISessions) []SessionID

	// Copy 复制一份（浅拷贝）
	Copy() IEnvelope
}

type Envelope struct {
	ID          string `json:"id" msgpack:"id"`
	AppID       string `json:"app_id" msgpack:"app_id"`
	Message     []byte `json:"message" msgpack:"message"`
	MessageType int    `json:"type" msgpack:"type"`
	Attempts    int    `json:"attempts" msgpack:"attempts"`

	// 由于消息可能来源于其他节点，为了保证日志的完整性，context需要记录一些trace信息
	Context string `json:"context" msgpack:"context"`
}

var _ IEnvelope = (*Envelope)(nil)

// newEnvelope，目前仅仅用于Close、Ping这2个原生的消息。
//
//	其他消息必须由继承者实现，不然调用Copy、GetExpectSessionIDs会panic
func newEnvelope(messageType int, message []byte) *Envelope {
	if messageType != CloseMessage && messageType != PingMessage {
		panic("invalid messageType")
	}
	return &Envelope{
		ID:          uuid.New().String(),
		MessageType: messageType,
		Message:     message,
		Attempts:    0,
	}
}

func (e *Envelope) GetMessageType() int {
	return e.MessageType
}

func (e *Envelope) GetMessage() []byte {
	return e.Message
}

func (e *Envelope) GetID() string {
	return e.ID
}

func (e *Envelope) GetAttempts() int {
	return e.Attempts
}

func (e *Envelope) GetOriginalAppID() string {
	return e.AppID
}

// GetExpectSessionIDs 获取期望的sessionID，如果返回nil，则表示不需要发送给任何session。由继承者实现
func (e *Envelope) GetExpectSessionIDs(sessions ISessions) []SessionID {
	panic("implement me")
}

// Copy 浅拷贝，复制一份。由继承者实现
func (e *Envelope) Copy() IEnvelope {
	panic("implement me")
}

func (e *Envelope) GetContext(parentCtx context.Context) context.Context {
	return utils.DeserializeContext(parentCtx, e.Context)
}

func (e *Envelope) SetMessageType(messageType int) {
	e.MessageType = messageType
}

func (e *Envelope) SetMessage(message []byte) {
	e.Message = message
}

func (e *Envelope) SetID(id string) {
	e.ID = id
}

func (e *Envelope) SetOriginalAppID(appID string) {
	e.AppID = appID
}

func (e *Envelope) SetAttempts(attempts int) {
	e.Attempts = attempts
}

func (e *Envelope) SetContext(context context.Context) {
	e.Context = utils.SerializeContext(context)
}

type EnvelopeBuilder[B any, E IEnvelope] struct {
	builder  B
	envelope E
}

func NewEnvelopeBuilder[B any, E IEnvelope](inheritedBuilder B, inheritedEnvelope E) *EnvelopeBuilder[B, E] {
	return &EnvelopeBuilder[B, E]{
		envelope: inheritedEnvelope,
		builder:  inheritedBuilder,
	}
}

func (b *EnvelopeBuilder[B, E]) WithID(id string) B {
	b.envelope.SetID(id)
	return b.builder
}

func (b *EnvelopeBuilder[B, E]) WithAppID(appID string) B {
	b.envelope.SetOriginalAppID(appID)
	return b.builder
}

func (b *EnvelopeBuilder[B, E]) WithAttempts(attempts int) B {
	b.envelope.SetAttempts(attempts)
	return b.builder
}

func (b *EnvelopeBuilder[B, E]) WithMessageType(messageType int) B {
	b.envelope.SetMessageType(messageType)
	return b.builder
}

func (b *EnvelopeBuilder[B, E]) WithMessage(message []byte) B {
	b.envelope.SetMessage(message)
	return b.builder
}

func (b *EnvelopeBuilder[B, E]) WithContext(context context.Context) B {
	b.envelope.SetContext(context)
	return b.builder
}

func (b *EnvelopeBuilder[B, E]) Build() E {
	if b.envelope.GetID() == "" {
		b.envelope.SetID(uuid.New().String())
	}
	if b.envelope.GetMessageType() == 0 {
		b.envelope.SetMessageType(TextMessage)
	}
	return b.envelope
}

type envelopeWrapper struct {
	Envelope  json.RawMessage `json:"envelope"`
	ClassName string          `json:"class_name"`
}

// MarshalIEnvelope 将IEnvelope转换为json
func MarshalIEnvelope(e IEnvelope) ([]byte, error) {
	var module string
	var bob json.RawMessage

	className := utils.GetClassName(e)
	if encoding, ok := envelopeCoding[className]; ok {
		bob, _ = encoding.Marshal(e)
		module = className
	} else {
		return nil, errors.New("invalid envelope")
	}

	return json.Marshal(&envelopeWrapper{
		Envelope:  bob,
		ClassName: module,
	})
}

// UnmarshalIEnvelope 将json转换为IEnvelope
func UnmarshalIEnvelope(data []byte) (IEnvelope, error) {
	var w envelopeWrapper
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}

	if encoding, ok := envelopeCoding[w.ClassName]; ok {
		return encoding.Unmarshal(w.Envelope)
	}

	return nil, errors.New("invalid envelope")
}
