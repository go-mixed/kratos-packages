package base

import "context"

// IEnvelope
// 切勿在SetXXX后面返回IEnvelope，因为返回的是*EnvelopeBase，而不是其继承的子集，会导致其继承的struct中的数据全丢失。
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

	// GetSendingConnections 通过conns中过滤出需要发送的 Connections，如果返回nil，则表示不需要发送给任何conn
	//  必须由继承者实现，根据自己的业务逻辑来返回，原始Envelope如果调用本方法，会panic
	GetSendingConnections(conns IConnections) IConnections

	// Copy 复制一份（浅拷贝）
	Copy() IEnvelope
}
