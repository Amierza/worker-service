package service

import (
	"context"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/helper"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/repository"
)

type (
	IAuthService interface {
		Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
		RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error)
	}

	authService struct {
		authRepo repository.IAuthRepository
		jwt      jwt.IJWT
	}
)

func NewAuthService(authRepo repository.IAuthRepository, jwt jwt.IJWT) *authService {
	return &authService{
		authRepo: authRepo,
		jwt:      jwt,
	}
}

func (as *authService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	if req.Password == "" || len(req.Password) < 8 {
		return dto.LoginResponse{}, dto.ErrInvalidPassword
	}

	user, found, err := as.authRepo.GetUserByIdentifier(ctx, nil, &req.Identifier)
	if err != nil {
		return dto.LoginResponse{}, dto.ErrGetUserByIdentifier
	}
	if !found {
		return dto.LoginResponse{}, dto.ErrUserNotFound
	}

	checkPassword, err := helper.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !checkPassword {
		return dto.LoginResponse{}, dto.ErrIncorrectPassword
	}

	accessToken, refreshToken, err := as.jwt.GenerateToken(user.ID.String(), string(user.Role))
	if err != nil {
		return dto.LoginResponse{}, dto.ErrGenerateAccessAndRefreshToken
	}

	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *authService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error) {
	_, err := as.jwt.ValidateToken(req.RefreshToken)
	if err != nil {
		return dto.RefreshTokenResponse{}, dto.ErrValidateToken
	}

	userID, err := as.jwt.GetUserIDByToken(req.RefreshToken)
	if err != nil {
		return dto.RefreshTokenResponse{}, dto.ErrGetUserIDFromToken
	}

	userRole, err := as.jwt.GetUserRoleByToken(req.RefreshToken)
	if err != nil {
		return dto.RefreshTokenResponse{}, dto.ErrGetUserRoleFromToken
	}

	accessToken, _, err := as.jwt.GenerateToken(userID, userRole)
	if err != nil {
		return dto.RefreshTokenResponse{}, dto.ErrGenerateAccessToken
	}

	return dto.RefreshTokenResponse{AccessToken: accessToken}, nil
}
