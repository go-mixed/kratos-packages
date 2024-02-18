package auth

import (
	"github.com/go-kratos/kratos/v2/errors"
)

const (
	// BearerWord the bearer key word for authorization
	BearerWord string = "Bearer"

	// AuthorizationKey holds the key used to store the token in the request tokenHeader.
	AuthorizationKey string = "Authorization"

	// UnauthorizedReason holds the error UnauthorizedReason.
	UnauthorizedReason string = "UNAUTHORIZED"
)

var (
	ErrMissingToken        = errors.Unauthorized(UnauthorizedReason, "token is missing")
	ErrWrongContext        = errors.Unauthorized(UnauthorizedReason, "Wrong context for middleware")
	ErrTokenExpired        = errors.Unauthorized(UnauthorizedReason, "token is expired")
	ErrTokenInvalid        = errors.Unauthorized(UnauthorizedReason, "token is invalid or disabled")
	ErrGuardNotFound       = errors.Unauthorized(UnauthorizedReason, "guard not found")
	ErrRefreshTokenInvalid = errors.Unauthorized(UnauthorizedReason, "refresh token is invalid or not found")
)
