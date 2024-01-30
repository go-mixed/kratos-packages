package auth

import (
	"github.com/go-kratos/kratos/v2/errors"
)

const (
	// BearerWord the bearer key word for authorization
	BearerWord string = "Bearer"

	// BearerFormat authorization token format
	BearerFormat string = "Bearer %s"

	// AuthorizationKey holds the key used to store the token in the request tokenHeader.
	AuthorizationKey string = "Authorization"

	// UnauthorizedReason holds the error UnauthorizedReason.
	UnauthorizedReason string = "UNAUTHORIZED"
)

var (
	ErrMissingToken = errors.Unauthorized(UnauthorizedReason, "token is missing")
	ErrWrongContext = errors.Unauthorized(UnauthorizedReason, "Wrong context for middleware")
)
