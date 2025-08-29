package dto

import (
	"errors"
)

const (
	// ====================================== Failed ======================================
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"

	// Token
	MESSAGE_FAILED_PROSES_REQUEST      = "failed proses request"
	MESSAGE_FAILED_ACCESS_DENIED       = "failed access denied"
	MESSAGE_FAILED_TOKEN_NOT_FOUND     = "failed token not found"
	MESSAGE_FAILED_TOKEN_NOT_VALID     = "failed token not valid"
	MESSAGE_FAILED_TOKEN_DENIED_ACCESS = "failed token denied access"

	// Authentication
	MESSAGE_FAILED_LOGIN_USER    = "failed login user"
	MESSAGE_FAILED_REFRESH_TOKEN = "failed refresh token"

	// ====================================== Success ======================================

	// Authentication
	MESSAGE_SUCCESS_LOGIN_USER    = "success login user"
	MESSAGE_SUCCESS_REFRESH_TOKEN = "success refresh token"
)

var (
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

	// Authentication
	ErrInvalidEmail      = errors.New("email must be in a valid format (ex: user123@example.com)")
	ErrInvalidPassword   = errors.New("password must be at least 8 characters long")
	ErrIncorrectPassword = errors.New("incorrect password")

	// User
	ErrGetUserByIdentifier = errors.New("failed get user by identifier")
	ErrUserNotFound        = errors.New("user not found")
)

// Authentiation for Admin
type (
	LoginRequest struct {
		Identifier string `json:"identifier" binding:"required" example:"187xxxxxx"`
		Password   string `json:"password" binding:"required" example:"secret123"`
	}
	LoginResponse struct {
		AccessToken  string `json:"access_token" example:"<access_token_here>"`
		RefreshToken string `json:"refresh_token" example:"<refresh_token_here>"`
	}
	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required" example:"<refresh_token_here>"`
	}
	RefreshTokenResponse struct {
		AccessToken string `json:"access_token" example:"<new_access_token_here>"`
	}
)
