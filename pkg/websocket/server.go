package websocket

import (
	"context"
	"crypto/tls"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/pkg/errors"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

type Server struct {
	*http.Server

	hub    base.IHub
	logger *log.Helper

	err      error
	listener net.Listener
	tlsConf  *tls.Config
	endpoint *url.URL
	upgrader *base.WsUpgrader
	network  string
	address  string
	path     string
}

// NewServer 实例化websocket
func NewServer(
	hub base.IHub,
	opts ...ServerOption) *Server {

	s := &Server{
		hub: hub,

		network: "tcp",
		address: ":0",
		path:    "/ws",
		upgrader: &base.WsUpgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		logger: log.NewModuleHelper(log.DefaultLogger, "websocket/server"),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.Server == nil {
		s.Server = &http.Server{
			TLSConfig: s.tlsConf,
		}
	}

	// 监听/ws的Http请求
	http.HandleFunc(s.path, s.ServeHTTP)

	s.err = s.listen()

	return s
}

// ServeHTTP upgrades http requests to websocket connections and dispatches them to be handled by the melody instance.
//
//	Http server的新建链接的入口，
//	注意：ServeHTTP是在http server的goroutine中执行的
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 将appID/RequestID放入request.Context中

	ctx := r.Context()
	ctx = requestid.NewContext(ctx, requestid.GetOrGenerateRequestId(r))

	logger := s.logger.WithContext(ctx)

	// recover panic
	defer func() {
		stackTrace := debug.Stack()
		stackTraceAsRawStringLiteral := strconv.Quote(string(stackTrace))
		if res := recover(); res != nil {
			logger.Errorf("[WS]ServeHTTP panic, request = %+v, recover = %+v,  stack = %s", r, res, stackTraceAsRawStringLiteral)
		}
	}()

	logger.Debugf("[WS]ServeHTTP request = %+v", *r)

	// 从http升级到websocket
	wsConn, err := s.upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		logger.Errorf("[WS]ServeHTTP upgrade error, request = %+v", r, err)
		s.responseError(w, http.StatusBadRequest, err)
		return
	}

	conn, err := s.hub.OnConnect(ctx, r, wsConn)
	if err != nil {
		_ = wsConn.WriteMessage(base.CloseMessage, []byte(err.Error()))
		_ = wsConn.Close()
		return
	}

	// conn.Close需要单独写一个defer，可以保证即使在其它defer中panic时，conn.Close也绝对会被执行。
	// 因为下文的Unregister、CallDisconnectHandler的链路太长，可能会panic
	defer func() {
		// 关闭conn
		conn.Close()
	}()

	// 离开函数时，反注册conn、关闭连接
	defer func() {
		if err = s.hub.OnDisconnect(conn); err != nil && !errors.Is(err, base.ErrServerClosed) {
			logger.Errorf("[WS]ServeHTTP onDisconnected error, request = %+v", r, err)
		}

	}()

	// [阻塞]死循环读取数据
	// 如果不阻塞，当前HTTP连接会断开
	conn.WaitForReceiving()
}

// responseError 返回http错误
func (s *Server) responseError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err.Error()))
}

// listen 监听server
func (s *Server) listen() error {
	if s.listener == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			return err
		}
		s.listener = lis
	}

	addr := s.address

	prefix := "ws://"
	if s.tlsConf == nil {
		if !strings.HasPrefix(addr, "ws://") {
			prefix = "ws://"
		}
	} else {
		if !strings.HasPrefix(addr, "wss://") {
			prefix = "wss://"
		}
	}
	addr = prefix + addr

	s.endpoint, s.err = url.Parse(addr)

	return nil
}

// Endpoint 实现endpoint接口
func (s *Server) Endpoint() (*url.URL, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

// Start 开启服务
func (s *Server) Start(ctx context.Context) error {
	logger := s.logger.WithContext(ctx)
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 64<<10)
			n := runtime.Stack(buf, false)
			buf = buf[:n]
			logger.Errorf("[WS]websocket panic error = %v, stack = %s", err, buf)
		}
	}()

	if s.err != nil {
		return s.err
	}
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	// 启动hub
	s.hub.OnServerStarted(ctx)

	s.logger.WithContext(ctx).Infof("[WS] server listening on: %s", s.listener.Addr().String())

	// 阻塞listen运行
	var err error
	if s.tlsConf != nil {
		err = s.ServeTLS(s.listener, "", "")
	} else {
		err = s.Serve(s.listener)
	}

	// 如果不是正常关闭的服务,则返回错误
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Stop 停止服务
func (s *Server) Stop(ctx context.Context) error {
	s.logger.WithContext(ctx).Info("[WS] server stopping")
	err := s.Shutdown(ctx)

	s.hub.OnServerStopped(ctx)

	return err
}
