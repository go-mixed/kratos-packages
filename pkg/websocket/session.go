package websocket

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Session wrapper around websocket connections.
type Session struct {
	ID       SessionID
	ActualID string
	Service  string
	Request  *http.Request

	conn   *Conn
	server *Server

	quitCh chan struct{} // 主动关闭session的channel

	open       atomic.Bool
	isObsolete bool // 被新的session替代了
	Data       *sync.Map
	ctx        context.Context
	fails      atomic.Uint32 // 读取、发送失败连续次数，只要成功发送、接收一次消息，就会重置为0

	lastRecvAt time.Time // 最近一次接收到消息的时间
	lastSendAt time.Time // 最近一次发送消息的时间

	mu sync.Mutex
}

func newSession(
	id SessionID,
	service string,
	conn *Conn,
) *Session {

	s := &Session{
		ID:       id,
		ActualID: fmt.Sprintf("%s:%s", time.Now().Format(time.RFC3339), uuid.New().String()),
		Service:  service,
		conn:     conn,

		quitCh: make(chan struct{}),

		open:       atomic.Bool{},
		isObsolete: false,
		Data:       &sync.Map{},
		ctx:        context.Background(),
		fails:      atomic.Uint32{},

		lastRecvAt: time.Now(), // 建立链接，就表示已经接收到了消息
		lastSendAt: time.Time{},

		mu: sync.Mutex{},
	}
	s.open.Store(true)
	return s
}

func (s *Session) initial(
	r *http.Request,
	server *Server,
	user auth.IGuard,
) {
	query := r.URL.Query()
	token := query.Get("token")
	version := query.Get("version")

	s.Request = r
	s.server = server
	s.ctx = r.Context()

	s.
		Set("id", s.ID).
		Set("ip", r.RemoteAddr).
		Set("token", token).
		Set("user", user).
		Set("service", s.Service).
		Set("requestID", requestid.FromContext(r.Context()))

	if version != "" {
		s.Set("version", version)
	}
}

func (s *Session) Context() context.Context {
	return s.ctx
}

func (s *Session) WithContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Session) updateLastRecvAt() {
	s.lastRecvAt = time.Now()
}

func (s *Session) updateLastSendAt() {
	s.lastSendAt = time.Now()
}

// Write a message to the websocket connection.
//
//	写数据到WS的conn中
func (s *Session) Write(envelope IEnvelope) error {
	// 无法并发写入：NextWriter、SetWriteDeadline、WriteMessage、WriteJSON、EnableWriteCompression、SetCompressionLevel；
	// 无法并发读取：NextReader、SetReadDeadline、ReadMessage、ReadJSON、SetPongHandler, SetPingHandler。
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	if s.Closed() {
		err = errors.Errorf("try to write to a closed session. session = %s. message = %+v", s, envelope)
		// 调用错误处理函数
		s.server.CallErrorHandler(s, err)
		return err
	}

	s.server.logger.Debugf("[WS]writing message to websocket. session = %s, message = %+v", s, envelope)

	// 每次写数据之前，先设置写超时时间
	if err = s.conn.SetWriteDeadline(time.Now().Add(s.server.WsConf.WriteTimeout)); err != nil {
		s.server.logger.Warn(errors.Wrapf(err, "SetWriteDeadline err. session = %s", s))
	}

	if err = s.conn.WriteMessage(envelope.GetMessageType(), envelope.GetMessage()); err != nil {
		// 错误次数+1
		s.fails.Add(1)
		err = errors.Wrapf(err, "WriteMessage err. session = %s", s)
		// 调用错误处理函数
		s.server.CallErrorHandler(s, err)
		return err
	}

	// 1次成功发送，就重置失败次数
	s.fails.Store(0)
	// 没有错误，更新最近一次发送消息的时间
	s.updateLastSendAt()

	return nil
}

// startPing sends a ping message to the client as a ticker
//
//	ping发送失败，不会尝试关闭session
func (s *Session) startPing() {
	if !s.Closed() {
		// 先创建下一个ping消息的定时器。
		// 为了记录fails的次数，此处无需主动Stop。
		// 等到下一次执行startPing时，会检查Closed，如果已经关闭，就不会再创建定时器了，所以不存在泄漏。
		time.AfterFunc(s.server.WsConf.PingInterval, s.startPing)

		e := newEnvelope(PingMessage, []byte("ping"))
		if err := s.Write(e); err != nil {
			s.server.logger.Warn(errors.Wrapf(err, "ping session %s failed", s))
		}
	}
}

//
//// sending pumps messages from the hub to the websocket connection.
//func (s *Session) sending() {
//	ticker := time.NewTicker(s.server.WsConf.PingInterval)
//	defer ticker.Stop()
//
//loop:
//	// 【阻塞】从sendCh中读取数据，写入到WS的conn中
//	for {
//		select {
//		case msg := <-s.sendCh:
//			// 任何写入错误（msg.ignoreError为true除外），都会导致session关闭。
//			// 错误包含：写入超时、链路断开
//			err := s.Write(msg)
//
//			if err != nil {
//				s.server.CallErrorHandler(s, errors.Wrapf(err, "sending message failed. session = %s", s))
//
//				// msg.ignoreError为false时，任何写入错误都会关闭session，并且退出循环
//				// ping/tryClose，直接走的doWriter不会关闭session，不会退出循环
//				if !msg.ignoreError {
//					_ = s.TryClose(nil)
//					break loop
//				}
//			}
//
//			// tryClose已经改为了doWrite，不会走到这里
//			// 如果手动writeToChannel(CloseMessage)，则会走到这里
//			if msg.t == CloseMessage {
//				s.server.logger.Debugf("[WS]sending close message. session = %s", s)
//				break loop
//			}
//
//			// doWrite已经更新最近一次发送消息的时间，此处调用handler
//			// ping/close消息不需要调用handler
//			if msg.t == TextMessage || msg.t == BinaryMessage {
//				_ = s.server.CallSendMessageHandler(s, msg.t, msg.msg)
//			}
//		case <-ticker.C:
//			s.ping() // 发送ping消息，走的是doWrite
//		case <-s.quitCh: // quitCh is closed when the session is closed
//			break loop
//		}
//	}
//}

// receiving pumps messages from the websocket connection to the hub.
func (s *Session) receiving() {
	// 设置conn的读取限制
	s.conn.SetReadLimit(s.server.WsConf.MaxMessageSize)

	// 设置conn的读取超时时间，因为Ping会在PongTimeout之前发送，接收到Pong之后，会延长读取下一个receive的超时时间。
	// 设置失败，也不影响程序的正常运行。
	if err := s.conn.SetReadDeadline(time.Now().Add(s.server.WsConf.PongTimeout)); err != nil {
		s.server.logger.Warn(errors.Wrapf(err, "receiving SetReadDeadline failed. session = %s", s))
	}

	// 启动ping定时任务
	s.startPing()

	// 设置conn的pong处理函数，回调pongHandler
	s.conn.SetPongHandler(func(string) error {
		// 只要收到pong，就重置失败次数
		s.fails.Store(0)
		// 延长读取下一个pong的超时时间，即使设置失败，也不影响程序的正常运行。
		// Pong的超时时间绝对要小于Ping的间隔时间
		if err := s.conn.SetReadDeadline(time.Now().Add(s.server.WsConf.PongTimeout)); err != nil {
			s.server.logger.Warn(errors.Wrapf(err, "SetPongHandler SetReadDeadline failed. session = %s", s))
		}
		// 先更新最近一次接收消息的时间，再调用handler
		s.updateLastRecvAt()
		_ = s.server.CallPongHandler(s)
		return nil
	})

	// 设置conn的关闭处理函数，来自于客户端的关闭，回调closeHandler
	s.conn.SetCloseHandler(func(code int, text string) error {
		s.server.logger.Warnf("[WS]client closed. session = %s", s)
		return s.server.CallCloseHandler(s, code, text)
	})

loop:
	// 【阻塞】读取conn的消息，这里不会收到pong消息，因为pong消息是在SetPongHandler中处理的
	for {
		// 超时、或服务器/客户端发起关闭，都会导致ReadMessage返回错误
		t, message, err := s.conn.ReadMessage()

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
				s.server.logger.Warn(errors.Wrapf(err, "receiving message failed. session = %s", s))
				break loop
			}

			// 错误次数+1
			s.fails.Add(1)
			s.server.CallErrorHandler(s, errors.Wrapf(err, "receiving message failed. session = %s", s))
			_ = s.TryClose("")
			break loop
		}

		// 1次成功接收，就重置失败次数
		s.fails.Store(0)
		// 先更新最近一次接收消息的时间，再调用handler
		s.updateLastRecvAt()
		_ = s.server.CallRecvMessageHandler(s, t, message)

		select {
		case <-s.quitCh: // quitCh is closed when the session is closed
			break loop
		default:

		}
	}
}

// TryClose try close session with a WS's Close-message.
//
//	这个方法会写一个CloseMessage到conn中
func (s *Session) TryClose(msg string) error {
	if !s.Closed() {
		return s.Write(newEnvelope(CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg)))
	}

	return errors.New("session is already closed during session.Close. session = " + s.String())
}

// Set is used to store a new key/value pair exclusivelly for this session.
// It also lazies initializes s.Data if it was not used previously.
func (s *Session) Set(key string, value any) *Session {
	s.Data.Store(key, value)
	return s
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (s *Session) Get(key string) (any, bool) {
	return s.Data.Load(key)
}

// Has returns true if the key exists.
func (s *Session) Has(key string) bool {
	_, ok := s.Data.Load(key)
	return ok
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (s *Session) MustGet(key string) any {
	if value, exists := s.Get(key); exists {
		return value
	}

	panic("Key \"" + key + "\" not exists in session = " + s.String())
}

// GetUser 返回当前session的用户
func (s *Session) GetUser() (auth.IAuth, error) {
	user := utils.SyncMapGet[auth.IAuth](s.Data, "user", nil)
	if user == nil || user.GetGuardName() == "" || user.GetGuardModel() == nil {
		return nil, errors.Errorf("user or guard is nil")
	}

	return user, nil
}

// GetToken 返回当前session的token
func (s *Session) GetToken() string {
	return utils.SyncMapGet[string](s.Data, "token", "")
}

func (s *Session) GetRemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// ClientIP implements one best effort algorithm to return the real client IP.
// it will try to parse and returns the headers defined in [X-Forwarded-For, X-Real-Ip].
// otherwise, the remote IP (coming from Request.RemoteAddr) is returned.
func (s *Session) ClientIP() string {
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

// Closed returns true if the session is closed.
//
//	在ServeHTTP结束后，open会设置为false
func (s *Session) Closed() bool {
	return !s.open.Load()
}

// Close closes the session and WS connection.
func (s *Session) Close() {
	open := s.open.Swap(false)

	if open {
		_ = s.conn.Close()
		close(s.quitCh)
		s.server.logger.Debugf("[WS]session connection closed. session = %s", s)
	}
}

func (s *Session) GetLastSendAt() time.Time {
	return s.lastSendAt
}

func (s *Session) GetLastRecvAt() time.Time {
	return s.lastRecvAt
}

func (s *Session) String() string {
	sb := &strings.Builder{}
	sb.WriteString("Session{")
	sb.WriteString("ID: ")
	sb.WriteString(s.ID.String())
	sb.WriteString(", ActualID: ")
	sb.WriteString(s.ActualID)
	sb.WriteString(", RemoteAddr: ")
	sb.WriteString(s.conn.RemoteAddr().String())
	sb.WriteString(", LastRecvAt: ")
	sb.WriteString(s.lastRecvAt.Format("2006-01-02 15:04:05"))
	sb.WriteString(", LastSendAt: ")
	sb.WriteString(s.lastSendAt.Format("2006-01-02 15:04:05"))

	user, _ := s.GetUser()
	if user != nil {
		sb.WriteString(", GuardName: ")
		sb.WriteString(user.GetGuardName())
		sb.WriteString(", GuardID: ")
		sb.WriteString(strconv.FormatInt(user.GetAuthorizationID(), 10))
	}

	sb.WriteString(", Token: ")
	sb.WriteString(s.GetToken())

	version, ok := s.Get("version")
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

// SetObsolete 标记当前session为过时的，也就是被新的session替换掉的
func (s *Session) SetObsolete() {
	s.isObsolete = true
}

// IsObsolete 如果当前session是被新的session替换掉的，那么就是true
func (s *Session) IsObsolete() bool {
	return s.isObsolete
}

// Fails 返回当前session的连续读取、写入失败次数
func (s *Session) Fails() int {
	return int(s.fails.Load())
}

// MakeSessionID 根据guard和service生成sessionID
func MakeSessionID(guard auth.IGuard, service string) SessionID {
	return SessionID(fmt.Sprintf("%s:%d:%s", guard.GetGuardName(), guard.GetAuthorizationID(), service))
}

// SentSessions 已经发送的sessions
type SentSessions struct {
	// 当前已经发送的sessions
	SentSessions ISessions
	// 期望发送的session ids
	ExpectSessionIDs []SessionID
	// 发送失败的session ids
	FailedSessionIDs []SessionID
}

// NeedAckSessions 在当前已发送的sessions中，找到需要ack的sessions
func (s *SentSessions) NeedAckSessions() ISessions {
	return s.SentSessions.Filter(FilterHasVersion())
}
