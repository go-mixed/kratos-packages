package hub

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/envelope"
)

func (s *Hub) GetConnections() base.IConnections {
	return s.connections
}

func (s *Hub) GetAllConnectionIDs(ctx context.Context) []base.ConnectionID {
	return s.connections.GetAllConnectionIDs(ctx)
}

// Send 根据messageType，发送消息
func (s *Hub) Send(ctx context.Context, messageType int, message []byte, connIds ...base.ConnectionID) error {
	if messageType == base.TextMessage {
		return s.SendText(ctx, string(message), connIds...)
	} else if messageType == base.BinaryMessage {
		return s.SendBinary(ctx, message, connIds...)
	}
	return errors.Errorf("[WS]message type %d not support", messageType)
}

// SendProtoMessage 根据messageType，将proto.Message转化为二进制、JSON之后发送。（支持Websocket连接不在当前节点）
func (s *Hub) SendProtoMessage(ctx context.Context, messageType int, message proto.Message, connIds ...base.ConnectionID) error {
	bytes := base.ProtoMarshal(messageType, message)
	if messageType == base.BinaryMessage {
		return s.SendBinary(ctx, bytes, connIds...)
	} else if messageType == base.TextMessage {
		return s.SendText(ctx, string(bytes), connIds...)
	}
	return errors.Errorf("[WS]message type %d not support", messageType)
}

// SendEnvelope 发送envelope。（支持Envelope中的conn ids不在当前节点）
func (s *Hub) SendEnvelope(ctx context.Context, envelope base.IEnvelope) error {
	// 先执行当前节点的发送
	err := s.sendRaw(ctx, envelope)
	// 发送到集群
	err = multierr.Append(err, s.cluster.Publish(ctx, envelope))
	return err
}

// SendText 发送text消息。（支持connIds不在当前节点）
func (s *Hub) SendText(ctx context.Context, message string, connIds ...base.ConnectionID) error {
	return s.SendEnvelope(ctx, envelope.NewEnvelopeBuilder().
		WithContext(ctx).
		WithMessageType(base.TextMessage).
		WithMessage([]byte(message)).
		WithConnectionIDs(false, connIds...).
		Build())
}

// SendBinary 发送binary消息。（支持connIds不在当前节点）
func (s *Hub) SendBinary(ctx context.Context, message []byte, connIds ...base.ConnectionID) error {
	return s.SendEnvelope(ctx, envelope.NewEnvelopeBuilder().
		WithContext(ctx).
		WithMessageType(base.BinaryMessage).
		WithMessage(message).
		WithConnectionIDs(false, connIds...).
		Build())
}

// BroadcastText 广播text消息。（会发送到所有节点）
func (s *Hub) BroadcastText(ctx context.Context, message string) error {
	return s.SendEnvelope(ctx, envelope.NewEnvelopeBuilder().
		WithContext(ctx).
		WithMessageType(base.TextMessage).
		WithMessage([]byte(message)).
		WithConnectionIDs(true).
		Build())
}

// BroadcastBinary 广播binary消息。（会发送到所有节点）
func (s *Hub) BroadcastBinary(ctx context.Context, message []byte) error {
	return s.SendEnvelope(ctx, envelope.NewEnvelopeBuilder().
		WithContext(ctx).
		WithMessageType(base.BinaryMessage).
		WithMessage(message).
		WithConnectionIDs(true).
		Build())
}

// Close 关闭连接，并发送exitMessage
func (s *Hub) Close(ctx context.Context, exitMessage string, connIds ...base.ConnectionID) {
	// 先关闭本地节点
	conns := s.connections.MGet(connIds...)
	for _, conn := range conns.Iterator() {
		_ = conn.TryClose(exitMessage)
	}

	// 发送关闭消息到集群
	_envelope := envelope.NewEnvelopeBuilder().
		WithContext(ctx).
		WithMessageType(base.CloseMessage).
		WithMessage(websocket.FormatCloseMessage(websocket.CloseNormalClosure, exitMessage)).
		WithConnectionIDs(false, connIds...).
		Build()

	_ = s.cluster.Publish(ctx, _envelope)
}

// sendRaw 仅仅发送本节点的消息，此函数也是集群中Pub/Sub的接收函数，用于处理envelope中位于本节点的消息
func (s *Hub) sendRaw(ctx context.Context, _envelope base.IEnvelope) error {
	if !s.Running() {
		return base.ErrServerClosed
	}

	if _envelope == nil {
		return nil
	}

	// 收到的是obsoleteEnvelope，交给专门的obsolete处理
	if e, isObsolete := _envelope.(*envelope.ObsoleteEnvelope); isObsolete {
		return s.doObsoleteConnection(e.ObsoleteConnectionID)
	}

	// 注意：Send的ctx可能来源于Http请求的request.ctx、或Publish。
	// 如果来源于HTTP请求，当退出本函数时，HTTP请求会结束，并且会cancel request.ctx
	// 因为Submit是异步函数，不会阻塞，所以需要复制一份ctx，避免在Submit异步任务中访问request.ctx
	// 使用一个被canceled的ctx在访问redis时，会报错，并且无法获取正确的结果，从而导致conns取值错误
	ctx = s.app.CloneContextFromBase(ctx)
	sentConns := s.doSendRaw(_envelope.GetContext(ctx), _envelope)
	if len(sentConns.FailedConnectionIDs) > 0 {
		return base.ErrOffline
	}

	return nil
}

// doSendRaw 真正发送消息的方法，只会发送本节点的消息
func (s *Hub) doSendRaw(ctx context.Context, _envelope base.IEnvelope) base.SentConnections {
	logger := s.logger.WithContext(ctx)
	var sentConns base.SentConnections

	if _envelope == nil {
		return sentConns
	}

	var sendingConns base.IConnections = _envelope.GetSendingConnections(s.GetConnections())
	var err error

	// 循环发送消息
	if sendingConns != nil {
		for _, conn := range sendingConns.Iterator() {
			// 写入错误时，尝试关闭conn
			if err = conn.Write(_envelope); err != nil {
				sentConns.FailedConnectionIDs = append(sentConns.FailedConnectionIDs, conn.GetID())
				// 达到失败阈值，关闭conn
				if conn.Fails() >= 0 {
					_ = conn.TryClose("")
				}
			} else {
				sentConns.SentConnectionIDs = append(sentConns.SentConnectionIDs, conn.GetID())
			}
		}
	}

	// 只有Text、Binary类型的消息，才会调用发送消息处理函数
	if _envelope.GetMessageType() == websocket.TextMessage ||
		_envelope.GetMessageType() == websocket.BinaryMessage {
		if err = s.callSendMessageHandler(ctx, _envelope, sentConns); err != nil {
			logger.Errorf("[WS]call send message handler error, envelope = %+v, connections = %v", _envelope, sendingConns.IDs(), err)
		}
	}

	return sentConns
}

// doObsoleteConnection 踢掉相同ID名的旧的conn，注意：是异步的
func (s *Hub) doObsoleteConnection(id base.ConnectionID) error {
	oldConn := s.GetConnections().Get(id)
	if oldConn != nil {
		oldConn.SetObsolete()
		return oldConn.TryClose("")
	}

	return nil
}
