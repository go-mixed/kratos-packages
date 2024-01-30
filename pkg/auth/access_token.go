package auth

import "time"

type IAccessToken interface {
	GetTokenableType() string
	GetTokenableId() int64
	GetAbilities() []string
	GetID() int64
	GetName() string
	ParseName(dst any) error
	GetToken() string
	//GetWrite() int
	GetLastUsedAt() time.Time
	GetCreatedAt() time.Time
	ExpiresAt(expiration time.Duration) time.Time
	ExpiresIn(expiration time.Duration) time.Duration
}
