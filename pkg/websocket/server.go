package websocket

import (
	"context"
	"crypto/tls"
	"github.com/go-kratos/kratos/v2/transport"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/grpc/encoding"

	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

type Server struct {
	*http.Server

	codec    encoding.Codec
	err      error
	logger   *log.Helper
	listener net.Listener
	tlsConf  *tls.Config
	endpoint *url.URL
	upgrader *Upgrader
	network  string
	address  string
	timeout  time.Duration
	path     string
	WsConf   *WsConfig

	handlers   []IWSHandler
	hub        IHub
	maxWorkers int
}

// NewServer 实例化websocket
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		WsConf:  defaultWsConfig(),
		network: "tcp",
		address: ":0",
		timeout: time.Second,
		path:    "/ws",
		upgrader: &Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		logger: log.NewModuleHelper(log.DefaultLogger, "websocket/server"),

		codec: encoding.GetCodec("json"),
		hub:   newEmptyHub(),
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

// RegisterHandlers 注册服务
func (s *Server) RegisterHandlers(handlers ...IWSHandler) {
	s.handlers = append(s.handlers, handlers...)
}

// CallConnectHandler calls the connect handler.
func (s *Server) CallConnectHandler(session *Session) error {
	for _, service := range s.handlers {
		if err := service.ConnectHandler(session.Context(), session); err != nil {
			return err
		}
	}
	return nil
}

// CallDisconnectHandler calls the disconnect handler.
func (s *Server) CallDisconnectHandler(session *Session) error {
	for _, service := range s.handlers {
		if err := service.DisconnectHandler(session.Context(), session); err != nil {
			return err
		}
	}
	return nil
}

// CallPongHandler calls the pong handler.
func (s *Server) CallPongHandler(session *Session) error {
	for _, service := range s.handlers {
		if err := service.PongHandler(session.Context(), session); err != nil {
			return err
		}
	}
	return nil
}

// CallRecvMessageHandler calls the recv message handler.
func (s *Server) CallRecvMessageHandler(session *Session, msgType int, msg []byte) error {
	for _, service := range s.handlers {
		if err := service.RecvMessageHandler(session.Context(), session, msgType, msg); err != nil {
			return err
		}
	}
	return nil
}

// CallSendMessageHandler calls the send message handler.
func (s *Server) CallSendMessageHandler(ctx context.Context, envelope IEnvelope, sentSessions SentSessions) error {
	for _, service := range s.handlers {
		if err := service.SendMessageHandler(ctx, envelope, sentSessions); err != nil {
			return err
		}
	}
	return nil
}

// CallErrorHandler calls the error handler.
func (s *Server) CallErrorHandler(session *Session, err error) {
	for _, service := range s.handlers {
		service.ErrorHandler(session.Context(), session, err)
	}
}

// CallCloseHandler calls the close handler.
func (s *Server) CallCloseHandler(session *Session, code int, text string) error {
	for _, service := range s.handlers {
		if err := service.CloseHandler(session.Context(), session, code, text); err != nil {
			return err
		}
	}
	return nil
}

// callStartHandler calls the start handler.
func (s *Server) callStartHandler(ctx context.Context) {
	for _, service := range s.handlers {
		service.StartHandler(ctx)
	}
}

// callStopHandler calls the stop handler.
func (s *Server) callStopHandler(ctx context.Context) {
	for _, service := range s.handlers {
		service.StopHandler(ctx)
	}
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
	conn, err := s.upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		logger.Errorf("[WS]ServeHTTP upgrade error, request = %+v", r, err)
		s.responseError(w, http.StatusBadRequest, err)
		_ = conn.Close()
		return
	}

	if s.hub.Closed() {
		logger.Error("[WS]ServeHTTP hub is closed")
		_ = conn.WriteMessage(CloseMessage, []byte("服务器正在维护中"))
		_ = conn.Close()
		return
	}

	accessToken, user, err := s.hub.Authenticate(r)
	if err != nil {
		logger.Errorf("[WS]ServeHTTP authenticate error, request = %+v", r, err)
		_ = conn.WriteMessage(CloseMessage, []byte("认证失败："+err.Error()))
		_ = conn.Close()
		return
	}

	session := s.makeSession(accessToken, user, r, conn)
	logger.Infof("[WS]connected, session = %s", session)

	// session.Close需要单独写一个defer，可以保证即使在其它defer中panic时，session.Close也绝对会被执行。
	// 因为下文的Unregister、CallDisconnectHandler的链路太长，可能会panic
	defer func() {
		// 关闭session
		session.Close()
	}()

	// 离开函数时，反注册session、关闭连接
	defer func() {
		if err = s.hub.Unregister(session); err != nil && !errors.Is(err, ErrHubClosed) {
			logger.Errorf("[WS]ServeHTTP unregister session error, request = %+v", r, err)
		}

		_ = s.CallDisconnectHandler(session)
	}()

	// 在hub中注册conn，返回session
	if err = s.hub.Register(session); err != nil {
		logger.Errorf("[WS]ServeHTTP register session error, request = %+v", r, err)
		_ = session.TryClose(err.Error())
		return
	}

	// 调用 connectHandler
	// 如果connectHandler返回err，那么直接返回，并断开连接
	if err = s.CallConnectHandler(session); err != nil {
		logger.Errorf("[WS]ServeHTTP call connectHandler error, request = %+v", r, err)
		_ = session.TryClose(err.Error())
		return
	}

	// [阻塞]死循环读取数据
	// 如果不阻塞，当前HTTP连接会断开
	session.receiving()
}

// responseError 返回http错误
func (s *Server) responseError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err.Error()))
}

// makeSession 创建session
func (s *Server) makeSession(accessToken auth.IAccessToken, auth auth.IGuard, r *http.Request, conn *Conn) *Session {
	query := r.URL.Query()
	service := query.Get("service")

	// 历史连接没有传递service参数
	if service == "" {
		service = ServiceDesktop
	}
	// 根据需求，当accessToken.name为2时，表示高级版
	if service == ServiceDesktop && strings.TrimSpace(accessToken.GetName()) == "2" {
		s.logger.WithContext(r.Context()).Infof("[WS]upgrade desktop to advanced, auth = %+v, accessToken = %+v", auth, accessToken)
		service = ServiceAdvancedDesktop
	}

	id := MakeSessionID(auth, service)
	session := newSession(id, service, conn)
	session.initial(r, s, auth)

	return session
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
	s.logger.WithContext(ctx).Infof("[WS]server listening on: %s", s.listener.Addr().String())

	// 启动hub
	s.hub.Start(ctx)

	// 调用startHandler，优先于广播、延迟队列启动
	s.callStartHandler(ctx)

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
	s.logger.WithContext(ctx).Info("[WS]server stopping")
	err := s.Shutdown(ctx)
	// 调用stopHandler
	s.callStopHandler(ctx)
	return err
}
