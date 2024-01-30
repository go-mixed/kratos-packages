package websocket

import (
	"time"

	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

type ServerOption func(o *Server)

func WithNetwork(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func WithAddress(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

func WithTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

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

// WithMaxWorkers 设置最大的worker数量，不设置则默认为cpu * 2
func WithMaxWorkers(num int) ServerOption {
	return func(o *Server) {
		o.maxWorkers = num
	}
}

func WithHub(hub IHub) ServerOption {
	return func(o *Server) {
		o.hub = hub
		hub.SetServer(o)
	}
}
