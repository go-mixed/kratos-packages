package websocket

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

const (
	MessageModuleDesktop = "desktop"
	MessageModulePms     = "pms"
	MessageModuleYi      = "yi"
)

const (
	ActionChatMessage = 10001
	ActionChatOnline  = 10002

	ActionReceptionMark = 20001

	ActionTerminalRzxAuth = 30001
)

type Conn = websocket.Conn
type Upgrader = websocket.Upgrader

type SessionID string

func (s SessionID) String() string {
	return string(s)
}

// for redis
func (s *SessionID) MarshalBinary() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *SessionID) UnmarshalBinary(data []byte) error {
	*s = SessionID(data)
	return nil
}

var ErrHubClosed = errors.New("hub closed")
var ErrInvalidEnvelope = errors.New("invalid envelope")
var ErrOfflineGuard = errors.New("offline guard")
