package websocket

import "time"

// WsConfig .
type WsConfig struct {
	// Milliseconds until write times out.
	WriteTimeout time.Duration
	// Timeout for waiting on pong.
	PongTimeout time.Duration
	// Milliseconds between pings.
	PingInterval time.Duration
	// Maximum size in bytes of a message.
	MaxMessageSize int64
	// The max amount of messages that can be in a sessions buffer before it starts dropping them.
	MessageBufferSize int
}

func defaultWsConfig() *WsConfig {
	return &WsConfig{
		WriteTimeout:      10 * time.Second,
		PongTimeout:       60 * time.Second,
		PingInterval:      (60 * time.Second * 9) / 10, // ping的间隔时间绝对要小于pong的超时时间
		MaxMessageSize:    512,
		MessageBufferSize: 256,
	}
}
