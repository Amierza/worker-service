package dto

import (
	"errors"
)

const (
	// ====================================== Failed ======================================
	// Token
	MESSAGE_FAILED_PROSES_REQUEST      = "failed proses request"
	MESSAGE_FAILED_ACCESS_DENIED       = "failed access denied"
	MESSAGE_FAILED_TOKEN_NOT_FOUND     = "failed token not found"
	MESSAGE_FAILED_TOKEN_NOT_VALID     = "failed token not valid"
	MESSAGE_FAILED_TOKEN_DENIED_ACCESS = "failed token denied access"

	// ====================================== Success ======================================
)

var (
	// Not Found
	ErrNotFound = errors.New("not found")
	// Unauthorized
	ErrUnauthorized = errors.New("unauthorized")

	// Token
	ErrGenerateAccessToken           = errors.New("failed to generate access token")
	ErrGenerateRefreshToken          = errors.New("failed to generate refresh token")
	ErrUnexpectedSigningMethod       = errors.New("unexpected signing method")
	ErrDecryptToken                  = errors.New("failed to decrypt token")
	ErrTokenInvalid                  = errors.New("token invalid")
	ErrValidateToken                 = errors.New("failed to validate token")
	ErrGetUserIDFromToken            = errors.New("failed get user id from token")
	ErrGetUserRoleFromToken          = errors.New("failed get user role from token")
	ErrGenerateAccessAndRefreshToken = errors.New("failed generate access and refresh token")
)
