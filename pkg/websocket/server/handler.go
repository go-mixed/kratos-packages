package server

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
)

// CallPongHandler calls the pong handler.
func (s *Server) CallPongHandler(conn base.IConnection) error {
	for _, handler := range s.handlers {
		if err := handler.PongHandler(conn.Context(), conn); err != nil {
			return err
		}
	}
	return nil
}

// CallRecvMessageHandler calls the recv message handler.
func (s *Server) CallRecvMessageHandler(conn base.IConnection, msgType int, msg []byte) error {
	for _, handler := range s.handlers {
		if err := handler.RecvMessageHandler(conn.Context(), conn, msgType, msg); err != nil {
			return err
		}
	}
	return nil
}

// CallSendMessageHandler calls the send message handler.
func (s *Server) CallSendMessageHandler(ctx context.Context, envelope base.IEnvelope, sentConns base.SentConnections) error {
	for _, handler := range s.handlers {
		if err := handler.SendMessageHandler(ctx, envelope, sentConns); err != nil {
			return err
		}
	}
	return nil
}

// CallErrorHandler calls the error handler.
func (s *Server) CallErrorHandler(conn base.IConnection, err error) {
	for _, handler := range s.handlers {
		handler.ErrorHandler(conn.Context(), conn, err)
	}
}

// CallCloseHandler calls the close handler.
func (s *Server) CallCloseHandler(conn base.IConnection, code int, text string) error {
	for _, handler := range s.handlers {
		if err := handler.CloseHandler(conn.Context(), conn, code, text); err != nil {
			return err
		}
	}
	return nil
}

// CallStartHandler calls the start handler.
func (s *Server) callStartHandler(ctx context.Context) {
	for _, handler := range s.handlers {
		handler.StartHandler(ctx)
	}
}

// CallStopHandler calls the stop handler.
func (s *Server) callStopHandler(ctx context.Context) {
	for _, handler := range s.handlers {
		handler.StopHandler(ctx)
	}
}

// callConnectHandler calls the connect handler.
func (s *Server) callConnectHandler(conn base.IConnection) error {
	for _, handler := range s.handlers {
		if err := handler.ConnectHandler(conn.Context(), conn); err != nil {
			return err
		}
	}
	return nil
}

// CallDisconnectHandler calls the disconnect handler.
func (s *Server) callDisconnectHandler(conn base.IConnection) error {
	for _, handler := range s.handlers {
		if err := handler.DisconnectHandler(conn.Context(), conn); err != nil {
			return err
		}
	}
	return nil
}
