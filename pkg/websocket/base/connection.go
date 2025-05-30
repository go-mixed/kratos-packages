package base

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"iter"
	"net"
	"time"
)

type ConnectionID string

func (s ConnectionID) String() string {
	return string(s)
}

// MarshalBinary for redis
func (s *ConnectionID) MarshalBinary() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *ConnectionID) UnmarshalBinary(data []byte) error {
	*s = ConnectionID(data)
	return nil
}

type IConnection interface {
	Context() context.Context
	WithContext(ctx context.Context)
	Write(envelope IEnvelope) error
	WaitForReceiving()

	TryClose(msg string) error

	SetMetadata(key string, value any)
	GetMetadata(key string) (any, bool)
	HasMetadata(key string) bool
	MustGetMetadata(key string) any

	GetUser() auth.IAuth
	GetRemoteAddr() net.Addr
	GetClientIP() string
	GetRequestId() string

	Closed() bool
	Close()

	GetLastSendAt() time.Time
	GetLastRecvAt() time.Time
	Fails() int

	SetObsolete()
	IsObsolete() bool

	String() string
	GetID() ConnectionID
}

type IConnections interface {
	Len() int
	Get(id ConnectionID) IConnection
	MGet(ids ...ConnectionID) IConnections
	HasKey(id ConnectionID) bool
	IDs() []ConnectionID
	Iterator(fns ...FilterConnectionFunc) iter.Seq2[ConnectionID, IConnection]
}

// SentConnections 已经发送的sessions
type SentConnections struct {
	// 当前已经发送的sessions
	SentConnectionIDs []ConnectionID
	// 发送失败的session ids
	FailedConnectionIDs []ConnectionID
}
