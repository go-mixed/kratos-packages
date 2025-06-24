package websocket

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

type ServerOption func(o *Server)

// WithNetwork 设置网络:tcp, udp, unix
func WithNetwork(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// WithAddress 设置监听地址: ip:port
func WithAddress(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// WithPath 监听路径，默认为 /ws
func WithPath(path string) ServerOption {
	return func(o *Server) {
		o.path = path
	}
}

// WithLogger 设置日志
func WithLogger(logger *log.Helper) ServerOption {
	return func(o *Server) {
		o.logger = logger
	}
}
