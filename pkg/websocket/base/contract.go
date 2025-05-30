package base

import (
	"errors"
	"github.com/gorilla/websocket"
)

const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)

type WsConn = websocket.Conn
type WsUpgrader = websocket.Upgrader

var ErrServerClosed = errors.New("websocket server closed")
var ErrInvalidEnvelope = errors.New("invalid envelope")
var ErrOffline = errors.New("websocket offline")
