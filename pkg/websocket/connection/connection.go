package connection

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/envelope"
	"iter"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Connection wrapper around websocket connections.
type Connection struct {
	ID      base.ConnectionID
	Request *http.Request

	conn         *base.WsConn
	handleCaller base.IHandleCaller
	user         auth.IAuth
	logger       *log.Helper
	conf         *base.WsConfig

	quitCh chan struct{} // 主动关闭connection的channel

	open       atomic.Bool
	isObsolete bool // 被新的connection替代了
	metadata   *utils.ConcurrentMap[string, any]
	ctx        context.Context
	fails      atomic.Uint32 // 读取、发送失败连续次数，只要成功发送、接收一次消息，就会重置为0

	lastRecvAt time.Time // 最近一次接收到消息的时间
	lastSendAt time.Time // 最近一次发送消息的时间
	createdAt  time.Time // 创建时间

	mu sync.Mutex
}

var _ base.IConnection = (*Connection)(nil)

func NewConnection(
	handleCaller base.IHandleCaller,
	r *http.Request,
	conn *base.WsConn,
	logger *log.Helper,
	conf *base.WsConfig,
	user auth.IAuth,
) base.IConnection {

	s := &Connection{
		ID:           base.ConnectionID(uuid.NewString()),
		conn:         conn,
		handleCaller: handleCaller,
		logger:       logger,

		quitCh: make(chan struct{}),

		open:       atomic.Bool{},
		isObsolete: false,
		metadata:   &utils.ConcurrentMap[string, any]{},
		ctx:        context.Background(),
		fails:      atomic.Uint32{},

		lastRecvAt: time.Now(), // 建立链接，就表示已经接收到了消息
		lastSendAt: time.Time{},
		createdAt:  time.Now(),
		conf:       conf,

		mu: sync.Mutex{},
	}
	s.open.Store(true)

	s.initial(r, user)
	return s
}

func (s *Connection) initial(
	r *http.Request,
	user auth.IAuth,
) {
	query := r.URL.Query()
	version := query.Get("version")

	s.Request = r
	s.user = user
	s.ctx = r.Context()

	if version != "" {
		s.SetMetadata("version", version)
	}

	// 将x-md-的头写入metadata中
	for k := range r.Header {
		k = strings.ToLower(k) // 小写
		if strings.HasPrefix(k, "x-md-") {
			s.SetMetadata(k, r.Header.Get(k))
		}
	}

}

func (s *Connection) Context() context.Context {
	return s.ctx
}

func (s *Connection) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Connection) touchLastRecvAt() {
	s.lastRecvAt = time.Now()
}

func (s *Connection) touchLastSendAt() {
	s.lastSendAt = time.Now()
}

// Write a message to the websocket connection.
//
//	写数据到WS的conn中
func (s *Connection) Write(envelope base.IEnvelope) error {
	// 无法并发写入：NextWriter、SetWriteDeadline、WriteMessage、WriteJSON、EnableWriteCompression、SetCompressionLevel；
	// 无法并发读取：NextReader、SetReadDeadline、ReadMessage、ReadJSON、SetPongHandler, SetPingHandler。
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	if s.IsClosed() {
		err = errors.Errorf("try to write to a closed connection. connection = %s. message = %+v", s, envelope)
		// 调用错误处理函数
		s.handleCaller.CallErrorHandler(s, err)
		return err
	}

	s.logger.Debugf("[WS]writing message to websocket. connection = %s, message = %+v", s, envelope)

	// 每次写数据之前，先设置写超时时间
	if err = s.conn.SetWriteDeadline(time.Now().Add(s.conf.WriteTimeout)); err != nil {
		s.logger.Warn(errors.Wrapf(err, "SetWriteDeadline err. connection = %s", s))
	}

	if err = s.conn.WriteMessage(envelope.GetMessageType(), envelope.GetMessage()); err != nil {
		// 错误次数+1
		s.fails.Add(1)
		err = errors.Wrapf(err, "WriteMessage err. connection = %s", s)
		// 调用错误处理函数
		s.handleCaller.CallErrorHandler(s, err)
		return err
	}

	// 1次成功发送，就重置失败次数
	s.fails.Store(0)
	// 没有错误，更新最近一次发送消息的时间
	s.touchLastSendAt()

	return nil
}

// startPing sends a ping message to the client as a ticker
//
//	ping发送失败，不会尝试关闭connection
func (s *Connection) startPing() {
	if !s.IsClosed() {
		// 先创建下一个ping消息的定时器。
		// 为了记录fails的次数，此处无需主动Stop。
		// 等到下一次执行startPing时，会检查Closed，如果已经关闭，就不会再创建定时器了，所以不存在泄漏。
		time.AfterFunc(s.conf.PingInterval, s.startPing)

		e := envelope.NewOperationEnvelope(base.PingMessage, []byte("ping"))
		if err := s.Write(e); err != nil {
			s.logger.Warn(errors.Wrapf(err, "ping connection %s failed", s))
		}
	}
}

//
//// sending pumps messages from the hub to the websocket connection.
//func (s *Connection) sending() {
//	ticker := time.NewTicker(s.handleCaller.wsConf.PingInterval)
//	defer ticker.Stop()
//
//loop:
//	// 【阻塞】从sendCh中读取数据，写入到WS的conn中
//	for {
//		select {
//		case msg := <-s.sendCh:
//			// 任何写入错误（msg.ignoreError为true除外），都会导致connection关闭。
//			// 错误包含：写入超时、链路断开
//			err := s.Write(msg)
//
//			if err != nil {
//				s.handleCaller.CallErrorHandler(s, errors.Wrapf(err, "sending message failed. connection = %s", s))
//
//				// msg.ignoreError为false时，任何写入错误都会关闭connection，并且退出循环
//				// ping/tryClose，直接走的doWriter不会关闭connection，不会退出循环
//				if !msg.ignoreError {
//					_ = s.TryClose(nil)
//					break loop
//				}
//			}
//
//			// tryClose已经改为了doWrite，不会走到这里
//			// 如果手动writeToChannel(CloseMessage)，则会走到这里
//			if msg.t == CloseMessage {
//				s.handleCaller.logger.Debugf("[WS]sending close message. connection = %s", s)
//				break loop
//			}
//
//			// doWrite已经更新最近一次发送消息的时间，此处调用handler
//			// ping/close消息不需要调用handler
//			if msg.t == TextMessage || msg.t == BinaryMessage {
//				_ = s.handleCaller.callSendMessageHandler(s, msg.t, msg.msg)
//			}
//		case <-ticker.C:
//			s.ping() // 发送ping消息，走的是doWrite
//		case <-s.quitCh: // quitCh is closed when the connection is closed
//			break loop
//		}
//	}
//}

func (s *Connection) cacheKey() string {
	return string("ws:connection:" + s.ID)
}

// WaitForReceiving pumps messages from the websocket connection to the hub, and block
func (s *Connection) WaitForReceiving() {

	// 设置conn的读取限制
	s.conn.SetReadLimit(s.conf.MaxMessageSize)

	// 设置conn的读取超时时间，因为Ping会在PongTimeout之前发送，接收到Pong之后，会延长读取下一个receive的超时时间。
	// 设置失败，也不影响程序的正常运行。
	if err := s.conn.SetReadDeadline(time.Now().Add(s.conf.PingTimeout)); err != nil {
		s.logger.Warn(errors.Wrapf(err, "receiving SetReadDeadline failed. connection = %s", s))
	}

	// 启动ping定时任务
	s.startPing()

	// 设置conn的pong处理函数，回调pongHandler
	s.conn.SetPongHandler(func(string) error {
		// 延长过期时间

		// 只要收到pong，就重置失败次数
		s.fails.Store(0)
		// 延长读取下一个pong的超时时间，即使设置失败，也不影响程序的正常运行。
		// Pong的超时时间绝对要小于Ping的间隔时间
		if err := s.conn.SetReadDeadline(time.Now().Add(s.conf.PingTimeout)); err != nil {
			s.logger.Warn(errors.Wrapf(err, "SetPongHandler SetReadDeadline failed. connection = %s", s))
		}
		// 先更新最近一次接收消息的时间，再调用handler
		s.touchLastRecvAt()
		_ = s.handleCaller.CallPongHandler(s)
		return nil
	})

	// 设置conn的关闭处理函数，来自于客户端的关闭，回调closeHandler
	s.conn.SetCloseHandler(func(code int, text string) error {
		s.logger.Warnf("[WS]client closed. connection = %s", s)
		return s.handleCaller.CallCloseHandler(s, code, text)
	})

loop:
	// 【阻塞】读取conn的消息，这里不会收到pong消息，因为pong消息是在SetPongHandler中处理的
	for {
		// 超时、或服务器/客户端发起关闭，都会导致ReadMessage返回错误
		typo, message, err := s.conn.ReadMessage()

		// 如果有错误，尝试优雅关闭
		// 然后退出循环
		// 客户端主动断开错误码是1005
		// CloseNormalClosure           = 1000 // 正常关闭
		// CloseGoingAway               = 1001 // 浏览器关闭
		// CloseProtocolError           = 1002 // 协议错误
		// CloseUnsupportedData         = 1003 // 接收到了不支持的数据类型
		// CloseNoStatusReceived        = 1005 // 接收到了客户端的关闭消息，但是没有关闭码
		// CloseAbnormalClosure         = 1006 // EOL
		// CloseInvalidFramePayloadData = 1007 // 接收到了无效的帧数据
		// ClosePolicyViolation         = 1008 // 接收到了不符合协议的数据
		// CloseMessageTooBig           = 1009 // 服务器/客户端发送的消息太大
		// CloseMandatoryExtension      = 1010 // 客户端需要扩展
		// CloseInternalServerErr       = 1011 // 服务器内部错误
		// CloseServiceRestart          = 1012 // 服务器重启
		// CloseTryAgainLater           = 1013 // 服务器临时不可用
		// CloseTLSHandshake            = 1015 // TLS握手失败
		if err != nil {
			// 客户端发起正常关闭，无需回调错误 https://github.com/Luka967/websocket-close-codes
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				s.logger.Warn(errors.Wrapf(err, "receiving message failed. connection = %s", s))
				break loop
			}

			// 错误次数+1
			s.fails.Add(1)
			s.handleCaller.CallErrorHandler(s, errors.Wrapf(err, "receiving message failed. connection = %s", s))
			_ = s.TryClose("")
			break loop
		}

		// 1次成功接收，就重置失败次数
		s.fails.Store(0)
		// 先更新最近一次接收消息的时间，再调用handler
		s.touchLastRecvAt()
		_ = s.handleCaller.CallRecvMessageHandler(s, typo, message)

		select {
		case <-s.quitCh: // quitCh is closed when the connection is closed
			break loop
		default:

		}
	}
}

// TryClose try close connection with a WS's Close-message.
//
//	这个方法会写一个CloseMessage到conn中
func (s *Connection) TryClose(msg string) error {
	if !s.IsClosed() {
		return s.Write(envelope.NewOperationEnvelope(base.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg)))
	}

	return errors.New("connection is already closed during connection.Close. connection = " + s.String())
}

// SetMetadata is used to store a new key/value pair exclusivelly for this connection.
// It also lazies initializes s.metadata if it was not used previously.
func (s *Connection) SetMetadata(key string, value any) {
	s.metadata.Store(key, value)
}

// GetMetadata returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (s *Connection) GetMetadata(key string) (any, bool) {
	return s.metadata.Load(key)
}

// HasMetadata returns true if the key exists.
func (s *Connection) HasMetadata(key string) bool {
	_, ok := s.metadata.Load(key)
	return ok
}

// MustGetMetadata returns the value for the given key if it exists, otherwise it panics.
func (s *Connection) MustGetMetadata(key string) any {
	if value, exists := s.GetMetadata(key); exists {
		return value
	}

	panic("Key \"" + key + "\" not exists in connection = " + s.String())
}

func (s *Connection) MetadataIterator() iter.Seq2[string, any] {
	return s.metadata.Iterator()
}

// GetID returns the connection ID.
func (s *Connection) GetID() base.ConnectionID {
	return s.ID
}

// SetID sets the connection ID.
func (s *Connection) SetID(id base.ConnectionID) {
	s.ID = id
}

// GetUser 返回当前connection的用户
func (s *Connection) GetUser() auth.IAuth {
	return s.user
}

func (s *Connection) GetRemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// GetClientIP implements one best effort algorithm to return the real client IP.
// it will try to parse and returns the headers defined in [X-Forwarded-For, X-Real-Ip].
// otherwise, the remote IP (coming from Request.RemoteAddr) is returned.
func (s *Connection) GetClientIP() string {
	removeAddr, _, _ := net.SplitHostPort(strings.TrimSpace(s.Request.RemoteAddr))
	remoteIP := net.ParseIP(removeAddr)
	if remoteIP == nil {
		return ""
	}
	// read real-ip from headers: X-Forwarded-For, X-Real-Ip
	remoteIPHeaders := []string{"X-Forwarded-For", "X-Real-IP"}
	for _, headerName := range remoteIPHeaders {
		header := s.Request.Header.Get(headerName)
		if header == "" {
			continue
		}

		// read the last ip
		ipStr, _ := lo.Last(strings.Split(header, ","))
		if ip := net.ParseIP(ipStr); ip == nil {
			break
		}
		return ipStr
	}

	return remoteIP.String()
}

// GetRequestId get the X-Request-Id from the request context
func (s *Connection) GetRequestId() string {
	return requestid.FromContext(s.ctx)
}

// GetRequest returns the request.
func (s *Connection) GetRequest() *http.Request {
	return s.Request
}

// IsClosed returns true if the connection is closed.
//
//	在ServeHTTP结束后，open会设置为false
func (s *Connection) IsClosed() bool {
	return !s.open.Load()
}

// Close closes the connection and WS connection.
func (s *Connection) Close() {
	open := s.open.Swap(false)

	if open {
		_ = s.conn.Close()
		close(s.quitCh)
		s.logger.Debugf("[WS]connection closed. connection = %s", s)
	}
}

func (s *Connection) GetLastSendAt() time.Time {
	return s.lastSendAt
}

func (s *Connection) GetLastRecvAt() time.Time {
	return s.lastRecvAt
}

func (s *Connection) String() string {
	sb := &strings.Builder{}
	sb.WriteString("Connection{")
	sb.WriteString("ID: ")
	sb.WriteString(s.ID.String())
	sb.WriteString(", Client IP: ")
	sb.WriteString(s.GetClientIP())
	sb.WriteString(", LastRecvAt: ")
	sb.WriteString(s.lastRecvAt.Format("2006-01-02 15:04:05"))
	sb.WriteString(", LastSendAt: ")
	sb.WriteString(s.lastSendAt.Format("2006-01-02 15:04:05"))

	user := s.GetUser()
	if user != nil {
		sb.WriteString(", GuardName: ")
		sb.WriteString(user.GetGuardModel().GetGuardName())
		sb.WriteString(", GuardID: ")
		sb.WriteString(strconv.FormatInt(user.GetGuardModel().GetAuthorizationID(), 10))
	}

	version, ok := s.GetMetadata("version")
	if ok {
		sb.WriteString(", Version: ")
		sb.WriteString(version.(string))
	}

	if s.IsObsolete() {
		sb.WriteString(", Obsolete: true")
	}

	sb.WriteString("}")
	return sb.String()
}

// SetObsolete 标记当前connection为过时的，也就是被新的connection替换掉的
func (s *Connection) SetObsolete() {
	s.isObsolete = true
}

// IsObsolete 如果当前connection是被新的connection替换掉的，那么就是true
func (s *Connection) IsObsolete() bool {
	return s.isObsolete
}

// Fails 返回当前connection的读取、写入失败次数（只要成功1次都会清零）
func (s *Connection) Fails() int {
	return int(s.fails.Load())
}
