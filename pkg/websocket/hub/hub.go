package hub

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/app"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	wsBase "gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/connection"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/envelope"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/service"
	"net/http"
	"sync/atomic"
)

type Hub struct {
	connections *connection.ConnectionManager
	handlers    []wsBase.IHandler
	adapter     wsBase.IAdapter
	grpcService *service.GrpcService

	cluster *wsCluster
	running atomic.Bool

	app    *app.App
	logger *log.Helper
	wsConf *wsBase.WsConfig
}

var _ wsBase.IHub = (*Hub)(nil)
var _ wsBase.IHandleCaller = (*Hub)(nil)

func NewHub(
	app *app.App,
	adapter wsBase.IAdapter,
	logger log.Logger,
	cache *cache.Cache,
	wsConf *wsBase.WsConfig,
) *Hub {
	s := &Hub{
		app:      app,
		adapter:  adapter,
		handlers: nil,
		running:  atomic.Bool{},

		connections: nil,
		logger:      log.NewModuleHelper(logger, "websocket/hub"),
		wsConf:      wsConf,
	}
	s.connections = connection.NewConnectionManager(cache, s.logger, s.wsConf.PingTimeout)
	s.cluster = newWsCluster(
		app,
		s.logger,
		cache.GetRedis(),
	)

	return s
}

func (s *Hub) Running() bool {
	return s.running.Load()
}

// RegisterHandlers 注册handler
func (s *Hub) RegisterHandlers(handlers ...wsBase.IHandler) {
	s.handlers = append(s.handlers, handlers...)
}

func (s *Hub) OnConnect(ctx context.Context, r *http.Request, wsConn *websocket.Conn) (wsBase.IConnection, error) {
	logger := s.logger.WithContext(ctx)

	if !s.Running() {
		logger.Error("[WS]ServeHTTP is closed")
		return nil, errors.New("服务器正在维护中")
	}

	user, err := s.adapter.Authorize(ctx, r)
	if err != nil {
		logger.Errorf("[WS]ServeHTTP authenticate error, request = %+v", r, err)
		return nil, errors.Errorf("认证失败：%s", err.Error())
	}

	// 将用户信息放入context中
	ctx = auth.NewContext(ctx, user)

	conn, err := s.makeConnection(ctx, r, wsConn, user)
	if err != nil {
		logger.Errorf("[WS]ServeHTTP invoke conn error, request = %+v", r, err)
		return nil, errors.Errorf("无法创建连接：%s", err.Error())
	}

	// 切勿将conn包裹到context中之后，再把ctx设置到conn中，这可能会造成无法gc
	// 此处只存储connection id，故不存在这个问题
	ctx = wsBase.NewContext(ctx, conn.GetID())
	conn.SetContext(ctx)

	logger = logger.WithContext(ctx)
	logger.Infof("[WS]connected, conn = %s", conn)

	// 调用 connectHandler
	if err = s.callConnectHandler(conn); err != nil {
		logger.Errorf("[WS]ServeHTTP call connectHandler error, request = %+v", r, err)
		return nil, err
	}

	s.upsertConnection(conn)

	logger.Infof("[WS]conn: \"%s\" connected", conn.GetID())
	return conn, nil
}

// makeConnection 创建conn
func (s *Hub) makeConnection(ctx context.Context, r *http.Request, wsConn *wsBase.WsConn, user auth.IAuth) (wsBase.IConnection, error) {
	// id := ConnectionID(fmt.Sprintf("%s:%d", auth.GetGuardName(), auth.GetAuthorizationID()))
	conn := connection.NewConnection(s, r, wsConn, s.logger, s.wsConf, user)

	var err error
	if conn, err = s.adapter.InvokeConnection(ctx, conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// upsertConnection 添加或更新连接。需要踢掉旧的conn
func (s *Hub) upsertConnection(conn wsBase.IConnection) {
	// 需要踢掉同名id的旧的conn，广播给其他节点，并设置踢它下线的conn的为Obsolete
	_ = s.doObsoleteConnection(conn.GetID())
	_ = s.cluster.Publish(conn.Context(), envelope.NewObsoleteEnvelope(conn.Context(), conn.GetID())) // 信封的ctx为新连接的ctx

	s.connections.Upsert(conn)
}

func (s *Hub) OnDisconnect(conn wsBase.IConnection) error {
	if !s.Running() {
		return wsBase.ErrServerClosed
	}

	// 在conns中删除，注意：必须是CompareAndDelete。当同名Id连接时，旧的conn会被挤掉，但是此时id不能被删除
	if s.connections.CompareAndDelete(conn.GetID(), conn) {
		s.logger.WithContext(conn.Context()).Infof("[WS]conn: \"%s\" disconncted", conn.GetID())
	}

	_ = s.callDisconnectHandler(conn)
	return nil
}

func (s *Hub) OnServerStarted(ctx context.Context) {
	s.running.Store(true)

	// 调用startHandler，优先于广播、延迟队列启动
	s.callStartHandler(ctx)

	// 监听其它服务器发来的信封
	s.cluster.Listen(ctx, s.sendRaw)
}

func (s *Hub) OnServerStopped(ctx context.Context) {
	s.running.Store(false)

	// 调用stopHandler
	s.callStopHandler(ctx)
}
