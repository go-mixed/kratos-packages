package auth

import (
	"github.com/go-kratos/kratos/v2/errors"
	"strings"
)

const (
	// BearerWord the bearer key word for authorization
	BearerWord string = "Bearer"

	// AuthorizationHeader holds the key used to store the token in the request tokenHeader.
	AuthorizationHeader string = "Authorization"

	// UnauthorizedReason holds the error UnauthorizedReason.
	UnauthorizedReason string = "UNAUTHORIZED"

	// ForbiddenReason holds the error ForbiddenReason.
	ForbiddenReason string = "FORBIDDEN"

	// ExpiredReason holds the error ExpiredReason.
	ExpiredReason string = "EXPIRED"

	LimitedReason string = "LIMITED"
)

var (
	ErrMissingToken        = errors.Unauthorized(UnauthorizedReason, "token is missing or invalid")
	ErrWrongContext        = errors.Unauthorized(UnauthorizedReason, "Wrong context for middleware")
	ErrTokenExpired        = errors.New(419, ExpiredReason, "token is expired")
	ErrRequestLimit        = errors.New(429, LimitedReason, "request count reaches the limit of the token")
	ErrTokenInvalid        = errors.Unauthorized(UnauthorizedReason, "token is invalid or disabled")
	ErrGuardNotFound       = errors.Unauthorized(UnauthorizedReason, "guard not found")
	ErrGuardNotMatch       = errors.Forbidden(ForbiddenReason, "the authorization guard not match the request guard")
	ErrTokenDisabled       = errors.Forbidden(ForbiddenReason, "access token is disabled")
	ErrRefreshTokenInvalid = errors.Unauthorized(UnauthorizedReason, "refresh token is invalid or not found")
	ErrForbidden           = errors.Forbidden(ForbiddenReason, "not allowed to access this resource")
)

// StripAuthorization strips the authorization prefix from the token.
func StripAuthorization(token string) string {
	authValue := strings.TrimSpace(token)
	if authValue == "" {
		return ""
	}

	// 从请求头中获取token，有Bearer开头的话去掉
	var requestToken string
	if !strings.HasPrefix(authValue, BearerWord) {
		requestToken = authValue
	} else {
		requestToken = strings.TrimSpace(authValue[len(BearerWord):])
	}

	return requestToken
}
