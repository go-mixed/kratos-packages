package hub

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
)

// CallPongHandler calls the pong handler.
func (s *Hub) CallPongHandler(conn base.IConnection) error {
	for _, handler := range s.handlers {
		if err := handler.OnPong(conn.Context(), conn); err != nil {
			return err
		}
	}
	return nil
}

// CallRecvMessageHandler calls the recv message handler.
func (s *Hub) CallRecvMessageHandler(connection base.IConnection, messageType int, msg []byte) error {
	ctx := connection.Context()
	_stream := base.NewStream(ctx, s, connection, messageType)
	ctx = base.NewContext(ctx, _stream)

	for _, handler := range s.handlers {
		if err := handler.OnRecvMessage(ctx, connection, messageType, msg); err != nil {
			return err
		}
	}
	return nil
}

// CallSendMessageHandler calls the send message handler.
func (s *Hub) callSendMessageHandler(ctx context.Context, envelope base.IEnvelope, sentConns base.SentConnections) error {
	for _, handler := range s.handlers {
		if err := handler.OnSendMessage(ctx, envelope, sentConns); err != nil {
			return err
		}
	}
	return nil
}

// CallErrorHandler calls the error handler.
func (s *Hub) CallErrorHandler(conn base.IConnection, err error) {
	for _, handler := range s.handlers {
		handler.OnError(conn.Context(), conn, err)
	}
}

// CallCloseHandler calls the close handler.
func (s *Hub) CallCloseHandler(conn base.IConnection, code int, text string) error {
	for _, handler := range s.handlers {
		if err := handler.OnClose(conn.Context(), conn, code, text); err != nil {
			return err
		}
	}
	return nil
}

// CallStartHandler calls the start handler.
func (s *Hub) callStartHandler(ctx context.Context) {
	for _, handler := range s.handlers {
		handler.OnStart(ctx)
	}
}

// CallStopHandler calls the stop handler.
func (s *Hub) callStopHandler(ctx context.Context) {
	for _, handler := range s.handlers {
		handler.OnStop(ctx)
	}
}

// callConnectHandler calls the connect handler.
func (s *Hub) callConnectHandler(conn base.IConnection) error {
	for _, handler := range s.handlers {
		if err := handler.OnConnect(conn.Context(), conn); err != nil {
			return err
		}
	}
	return nil
}

// CallDisconnectHandler calls the disconnect handler.
func (s *Hub) callDisconnectHandler(conn base.IConnection) error {
	for _, handler := range s.handlers {
		if err := handler.OnDisconnect(conn.Context(), conn); err != nil {
			return err
		}
	}
	return nil
}
