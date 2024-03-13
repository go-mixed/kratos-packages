package auth

import (
	"github.com/go-kratos/kratos/v2/errors"
	"net/http"
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
)

var (
	ErrMissingToken        = errors.Unauthorized(UnauthorizedReason, "token is missing")
	ErrWrongContext        = errors.Unauthorized(UnauthorizedReason, "Wrong context for middleware")
	ErrTokenExpired        = errors.New(http.StatusNotAcceptable, ExpiredReason, "token is expired")
	ErrTokenInvalid        = errors.Unauthorized(UnauthorizedReason, "token is invalid or disabled")
	ErrGuardNotFound       = errors.Unauthorized(UnauthorizedReason, "guard not found")
	ErrGuardNotMatch       = errors.Forbidden(ForbiddenReason, "the authorization guard not match the request guard")
	ErrTokenDisabled       = errors.Forbidden(ForbiddenReason, "access token is disabled")
	ErrRefreshTokenInvalid = errors.Unauthorized(UnauthorizedReason, "refresh token is invalid or not found")
	ErrForbidden           = errors.Forbidden(ForbiddenReason, "not allowed to access this resource")
)
