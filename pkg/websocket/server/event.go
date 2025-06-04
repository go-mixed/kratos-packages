package server

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/envelope"
)

func (s *Server) onServerStarted(ctx context.Context) {

	s.running.Store(true)

	// 调用startHandler，优先于广播、延迟队列启动
	s.callStartHandler(ctx)

	// 监听信封
	s.cluster.Listen(ctx, s.sendRaw)

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
