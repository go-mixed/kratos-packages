package server

import (
	"context"
	"crypto/tls"
	"github.com/go-kratos/kratos/v2/transport"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/connection"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/service"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
	_ base.IHub            = (*Server)(nil)
	_ base.IHandleCaller   = (*Server)(nil)
)

type Server struct {
	*http.Server

	app     *app.App
	logger  *log.Helper
	running atomic.Bool

	connections *connection.ConnectionManager
	handlers    []base.IHandler
	adapter     base.IAdapter
	grpcService *service.GrpcService

	cluster *wsCluster

	err      error
	listener net.Listener
	tlsConf  *tls.Config
	endpoint *url.URL
	upgrader *base.WsUpgrader
	network  string
	address  string
	timeout  time.Duration
	path     string
	wsConf   *base.WsConfig
}

// NewServer 实例化websocket
func NewServer(
	app *app.App,
	cache *cache.Cache,
	adapter base.IAdapter,

	opts ...ServerOption) *Server {

	s := &Server{
		app:      app,
		adapter:  adapter,
		handlers: nil,

		connections: nil,
		running:     atomic.Bool{},

		wsConf:  base.DefaultWsConfig(),
		network: "tcp",
		address: ":0",
		timeout: time.Second,
		path:    "/ws",
		upgrader: &base.WsUpgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		logger: log.NewModuleHelper(log.DefaultLogger, "websocket/handleCaller"),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.connections = connection.NewConnectionManager(cache, s.logger, s.wsConf.PongTimeout)
	s.cluster = newWsCluster(
		app,
		s.logger,
		cache.GetRedis(),
	)

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

// RegisterHandlers 注册handler
func (s *Server) RegisterHandlers(handlers ...base.IHandler) {
	s.handlers = append(s.handlers, handlers...)
}

// ServeHTTP upgrades http requests to websocket connections and dispatches them to be handled by the melody instance.
//
//	Http server的新建链接的入口，
//	注意：ServeHTTP是在http server的goroutine中执行的
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 将appID/RequestID放入request.Context中

	ctx := r.Context()
	ctx = requestid.NewContext(ctx, requestid.GetOrGenerateRequestId(r))
	r = r.WithContext(ctx)

	logger := s.logger.WithContext(ctx)

	// recover panic
	defer func() {
		if res := recover(); res != nil {
			logger.Errorf("[WS]ServeHTTP panic, request = %+v, recover = %+v", r, res)
		}
	}()

	logger.Debugf("[WS]ServeHTTP request = %+v", *r)

	// 从http升级到websocket
	wsConn, err := s.upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		logger.Errorf("[WS]ServeHTTP upgrade error, request = %+v", r, err)
		s.responseError(w, http.StatusBadRequest, err)
		_ = wsConn.Close()
		return
	}

	if !s.Running() {
		logger.Error("[WS]ServeHTTP is closed")
		_ = wsConn.WriteMessage(base.CloseMessage, []byte("服务器正在维护中"))
		_ = wsConn.Close()
		return
	}

	user, err := s.adapter.Authorize(ctx, r)
	if err != nil {
		logger.Errorf("[WS]ServeHTTP authenticate error, request = %+v", r, err)
		_ = wsConn.WriteMessage(base.CloseMessage, []byte("认证失败："+err.Error()))
		_ = wsConn.Close()
		return
	}

	conn, err := s.makeConnection(ctx, r, wsConn, user)
	if err != nil {
		logger.Errorf("[WS]ServeHTTP invoke conn error, request = %+v", r, err)
		_ = wsConn.WriteMessage(base.CloseMessage, []byte("无法创建会话："+err.Error()))
		_ = wsConn.Close()
		return
	}

	logger.Infof("[WS]connected, conn = %s", conn)

	// conn.Close需要单独写一个defer，可以保证即使在其它defer中panic时，conn.Close也绝对会被执行。
	// 因为下文的Unregister、CallDisconnectHandler的链路太长，可能会panic
	defer func() {
		// 关闭conn
		conn.Close()
	}()

	// 离开函数时，反注册conn、关闭连接
	defer func() {
		if err = s.onDisconnected(conn); err != nil && !errors.Is(err, base.ErrServerClosed) {
			logger.Errorf("[WS]ServeHTTP unregister conn error, request = %+v", r, err)
		}

		_ = s.callDisconnectHandler(conn)
	}()

	// 在hub中注册conn，返回conn
	if err = s.onConnected(conn); err != nil {
		logger.Errorf("[WS]ServeHTTP register conn error, request = %+v", r, err)
		_ = conn.TryClose(err.Error())
		return
	}

	// 调用 connectHandler
	// 如果connectHandler返回err，那么直接返回，并断开连接
	if err = s.callConnectHandler(conn); err != nil {
		logger.Errorf("[WS]ServeHTTP call connectHandler error, request = %+v", r, err)
		_ = conn.TryClose(err.Error())
		return
	}

	// [阻塞]死循环读取数据
	// 如果不阻塞，当前HTTP连接会断开
	conn.WaitForReceiving()
}

func (s *Server) Running() bool {
	return s.running.Load()
}

// responseError 返回http错误
func (s *Server) responseError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err.Error()))
}

// makeConnection 创建conn
func (s *Server) makeConnection(ctx context.Context, r *http.Request, wsConn *base.WsConn, user auth.IAuth) (base.IConnection, error) {
	// id := ConnectionID(fmt.Sprintf("%s:%d", auth.GetGuardName(), auth.GetAuthorizationID()))
	conn := connection.NewConnection(s, r, wsConn, s.logger, s.wsConf, user)

	var err error
	if conn, err = s.adapter.InvokeConnection(ctx, conn); err != nil {
		return nil, err
	}

	return conn, nil
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
	s.logger.WithContext(ctx).Infof("[WS]handleCaller listening on: %s", s.listener.Addr().String())

	// 启动hub
	s.onServerStarted(ctx)

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
	s.logger.WithContext(ctx).Info("[WS]handleCaller stopping")
	err := s.Shutdown(ctx)

	s.onServerStopped(ctx)

	return err
}
