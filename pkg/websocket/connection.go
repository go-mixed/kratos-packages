package websocket

import (
	"encoding"
	"encoding/json"
	"github.com/samber/lo"
	"time"
)

type Connection struct {
	GuardName string `json:"guard_name" redis:"guard_name"`
	GuardId   int64  `json:"guard_id" redis:"guard_id"`

	AppID           string    `json:"app_id" redis:"app_id"`
	SessionID       SessionID `json:"session_id" redis:"session_id"`
	SessionActualID string    `json:"session_actual_id" redis:"session_actual_id"`
	Token           string    `json:"token" redis:"token"`
	Service         string    `json:"service" redis:"service"`
	RemoteAddr      string    `json:"remote_addr" redis:"remote_addr"`
	ClientIP        string    `json:"client_ip" redis:"client_ip"`

	LastSendAt time.Time `json:"last_send_at" redis:"last_send_at" `
	LastRecvAt time.Time `json:"last_recv_at" redis:"last_recv_at"`
	OfflineAt  time.Time `json:"offline_at" redis:"offline_at"`
	IsObsolete bool      `json:"is_obsolete" redis:"is_obsolete"`
}

var _ encoding.BinaryMarshaler = (*Connection)(nil)
var _ encoding.BinaryUnmarshaler = (*Connection)(nil)

func NewConnection(appID string, session *Session) *Connection {
	conn := &Connection{
		SessionID:       session.ID,
		SessionActualID: session.ActualID,
		Token:           session.GetToken(),
		AppID:           appID,
		Service:         session.Service,
		RemoteAddr:      session.GetRemoteAddr().String(),
		ClientIP:        session.ClientIP(),
		LastSendAt:      session.GetLastSendAt(),
		LastRecvAt:      session.GetLastRecvAt(),
		IsObsolete:      session.IsObsolete(),
	}

	user, _ := session.GetUser()
	if user != nil {
		conn.GuardName = user.GetGuardName()
		conn.GuardId = user.GetAuthorizationID()
	}

	return conn
}

func (c *Connection) LastUpdateAt() time.Time {
	if c.LastSendAt.After(c.LastRecvAt) {
		return c.LastSendAt
	}
	return c.LastRecvAt
}

// MarshalBinary 传递给redis的数据必须为指针（比如：Set('key', &Connection{} )，不然本函数无法满足BinaryMarshaler接口
func (c *Connection) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

func (c *Connection) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

type connections[K comparable] map[K]*Connection

type Connections = connections[SessionID]

type ActualConnections = connections[string]

func (c connections[K]) Get(id K) *Connection {
	return c[id]
}

func (c connections[K]) Set(id K, conn *Connection) connections[K] {
	c[id] = conn
	return c
}

func (c connections[K]) Delete(id K) connections[K] {
	delete(c, id)
	return c
}

// FilterBefore 返回所有最近一次更新时间在beforeAt之前（不含）的连接
func (c connections[K]) FilterBefore(beforeAt time.Time) connections[K] {
	return lo.PickBy(c, func(_ K, conn *Connection) bool {
		return conn.LastUpdateAt().Before(beforeAt)
	})
}

// FilterAfter 返回所有最近一次更新时间在afterAt之后（不含）的连接
func (c connections[K]) FilterAfter(afterAt time.Time) connections[K] {
	return lo.PickBy(c, func(_ K, conn *Connection) bool {
		return conn.LastUpdateAt().After(afterAt)
	})
}

// FilterApp 返回appID的所有连接
func (c connections[K]) FilterApp(appID string) connections[K] {
	return lo.PickBy(c, func(_ K, conn *Connection) bool {
		return conn.AppID == appID
	})
}

// FilterService 匹配service，返回所匹配的连接
func (c connections[K]) FilterService(service string) connections[K] {
	return lo.PickBy(c, func(_ K, conn *Connection) bool {
		return conn.Service == service
	})
}

// FilterAnyServices 匹配services中任意一个，返回所匹配的连接
func (c connections[K]) FilterAnyServices(services ...string) connections[K] {
	if len(services) == 0 {
		return c
	}
	return lo.PickBy(c, func(_ K, conn *Connection) bool {
		return lo.Contains(services, conn.Service)
	})
}

// SessionIDs 返回当前Connections中所有的sessionID
func (c connections[K]) SessionIDs() []SessionID {
	return lo.MapToSlice(c, func(_ K, value *Connection) SessionID {
		return value.SessionID
	})
}

// SessionActualID 返回当前Connections中所有的sessionActualID
func (c connections[K]) SessionActualID() []string {
	return lo.MapToSlice(c, func(_ K, value *Connection) string {
		return value.SessionActualID
	})
}

// Services 返回当前Connections中所有的service
func (c connections[K]) Services() []string {
	return lo.MapToSlice(c, func(_ K, value *Connection) string {
		return value.Service
	})
}
