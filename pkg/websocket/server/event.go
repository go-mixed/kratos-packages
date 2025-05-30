package server

import (
	"context"
	"github.com/gorilla/websocket"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/envelope"
)

func (s *Server) onServerStarted(ctx context.Context) {
	// 监听信封
	s.cluster.Listen(ctx, s.sendRaw)

	s.running.Store(true)

	// 调用startHandler，优先于广播、延迟队列启动
	s.callStartHandler(ctx)
}

func (s *Server) onServerStopped(ctx context.Context) {
	s.running.Store(false)
	// 调用stopHandler
	s.callStopHandler(ctx)
}

func (s *Server) onConnected(conn base.IConnection) error {
	if !s.Running() {
		return base.ErrServerClosed
	}

	// 需要踢掉同名id的旧的conn，广播给其他节点，并设置踢它下线的conn的为Obsolete
	_ = s.doObsoleteConnection(conn.GetID())
	_ = s.cluster.Publish(conn.Context(), envelope.NewObsoleteEnvelope(conn.Context(), conn.GetID())) // 信封的ctx为新连接的ctx

	s.connections.Upsert(conn)
	s.logger.WithContext(conn.Context()).Infof("[WS]conn: \"%s\" connected", conn.GetID())

	return nil
}

func (s *Server) onDisconnected(conn base.IConnection) error {
	if !s.Running() {
		return base.ErrServerClosed
	}

	// 在conns中删除，注意：必须是CompareAndDelete，当同名Id连接时，旧的conn会被挤掉，但是此时id不能被删除
	if s.connections.CompareAndDelete(conn.GetID(), conn) {
		s.logger.WithContext(conn.Context()).Infof("[WS]conn: \"%s\" disconncted", conn.GetID())
	}
	return nil
}

func (s *Server) sendRaw(ctx context.Context, _envelope base.IEnvelope) error {
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

// doSendRaw 真正发送消息的方法
//  1. 获取当前需要发送的connIDs
//  2. 根据redis中的envelope zset，过滤掉已经ack成功的connIDs
//  3. 如果是重试消息，需要过滤掉无需重试的connIDs，即存在version才需要重试
//  4. 循环发送消息
func (s *Server) doSendRaw(ctx context.Context, _envelope base.IEnvelope) base.SentConnections {
	logger := s.logger.WithContext(ctx)
	var sentConns base.SentConnections

	if _envelope == nil {
		return sentConns
	}

	var sendingConns base.IConnections = _envelope.GetSendingConnections(s.connections)
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
		if err = s.CallSendMessageHandler(ctx, _envelope, sentConns); err != nil {
			logger.Errorf("[Hub]call send message handler error, envelope = %+v, connections = %v", _envelope, sendingConns.IDs(), err)
		}
	}

	return sentConns
}

// doObsoleteConnection 踢掉相同ID名的旧的conn，注意：是异步的
func (s *Server) doObsoleteConnection(id base.ConnectionID) error {
	oldConn := s.connections.Get(id)
	if oldConn != nil {
		oldConn.SetObsolete()
		return oldConn.TryClose("")
	}

	return nil
}
