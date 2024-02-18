package auth

import "time"

type IAccessToken interface {
	GetID() int64
	GetAbilities() []string
	SetAbilities(val []string)
	GetRefreshToken() string
	GetLastUsedAt() time.Time
	GetCreatedAt() time.Time
	GetExpiredAt() time.Time
	SetExpiredAt(val time.Time)
	IsEnabled() bool
	IGuard
}
