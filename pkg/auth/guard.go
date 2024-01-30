package auth

import (
	"encoding/json"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
)

type IGuard interface {
	GetGuardName() string
	GetAuthorizationID() int64
}

type IGuardModel interface {
	IGuard
	db.IMorphTabler
}

const GuardShop = "shop"
const GuardMerchant = "merchant"
const GuardStaff = "staff"
const GuardChatUser = "chat_user"
const GuardTerminal = "terminal"

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

// GuardToTokenType 将guardName转换为token_type（即PHPackage）
func GuardToTokenType(guard string) string {
	switch guard {
	case GuardShop:
		return "App\\Models\\Shop"
	case GuardMerchant:
		return "App\\Models\\Merchant"
	case GuardStaff:
		return "App\\Models\\Staff"
	case GuardChatUser:
		return "App\\Models\\Chat\\ChatUser"
	case GuardTerminal:
		return "App\\Models\\ShopTerminal"
	default:
		return ""
	}
}

// TokenTypeToGuard 将token_type（即PHPackage）转换为guardName
func TokenTypeToGuard(tokenType string) string {
	switch tokenType {
	case "App\\Models\\Shop":
		return GuardShop
	case "App\\Models\\Merchant":
		return GuardMerchant
	case "App\\Models\\Staff":
		return GuardStaff
	case "App\\Models\\Chat\\ChatUser":
		return GuardChatUser
	case "App\\Models\\ShopTerminal":
		return GuardTerminal
	default:
		return ""
	}
}
