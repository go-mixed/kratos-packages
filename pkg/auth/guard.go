package auth

import (
	"encoding/json"
)

type IGuard interface {
	GetGuardName() string
	GetAuthorizationID() int64
}

type Guard struct {
	GuardName       string `json:"guard_name" msgpack:"guard_name" redis:"guard_name"`
	AuthorizationID int64  `json:"authorization_id" msgpack:"authorization_id" redis:"authorization_id"`
}

var _ IGuard = (*Guard)(nil)

// NewGuard 创建一个新的Guard
func NewGuard(guardName string, authorizationID int64) *Guard {
	return &Guard{
		GuardName:       guardName,
		AuthorizationID: authorizationID,
	}
}

// WrapGuard 使用IGuard包装一个Guard
func WrapGuard(guard IGuard) *Guard {
	if guard == nil {
		return nil
	} else if g, ok := guard.(*Guard); ok {
		return g
	}

	return NewGuard(guard.GetGuardName(), guard.GetAuthorizationID())
}

func (g *Guard) GetGuardName() string {
	return g.GuardName
}

func (g *Guard) GetAuthorizationID() int64 {
	return g.AuthorizationID
}

func (g *Guard) MarshalBinary() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Guard) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, g)
}
