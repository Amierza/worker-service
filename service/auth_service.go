package service

import (
	"context"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/helper"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/repository"
	"go.uber.org/zap"
)

type (
	IAuthService interface {
		Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
		RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error)
	}

	authService struct {
		authRepo repository.IAuthRepository
		logger   *zap.Logger
		jwt      jwt.IJWT
	}
)

func NewAuthService(authRepo repository.IAuthRepository, logger *zap.Logger, jwt jwt.IJWT) *authService {
	return &authService{
		authRepo: authRepo,
		logger:   logger,
		jwt:      jwt,
	}
}

func (as *authService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	user, found, err := as.authRepo.GetUserByIdentifier(ctx, nil, &req.Identifier)
	if err != nil {
		as.logger.Error("failed to get user by identifier",
			zap.String("identifier", req.Identifier),
			zap.Error(err),
		)
		return dto.LoginResponse{}, dto.ErrGetUserByIdentifier
	}
	if !found {
		as.logger.Warn("user not found",
			zap.String("identifier", req.Identifier),
		)
		return dto.LoginResponse{}, dto.ErrNotFound
	}

	checkPassword, err := helper.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !checkPassword {
		as.logger.Error("password not match",
			zap.String("password", req.Password),
			zap.Error(err),
		)
		return dto.LoginResponse{}, dto.ErrIncorrectPassword
	}

	accessToken, refreshToken, err := as.jwt.GenerateToken(user.ID.String(), string(user.Role))
	if err != nil {
		as.logger.Error("failed to generate access and refresh token",
			zap.String("identifier", req.Identifier),
			zap.Error(err),
		)
		return dto.LoginResponse{}, dto.ErrGenerateAccessAndRefreshToken
	}

	as.logger.Info("login success",
		zap.String("identifier", req.Identifier),
	)

	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *authService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error) {
	_, err := as.jwt.ValidateToken(req.RefreshToken)
	if err != nil {
		as.logger.Error("invalid token",
			zap.String("refresh_token", req.RefreshToken),
			zap.Error(err),
		)
		return dto.RefreshTokenResponse{}, dto.ErrValidateToken
	}

	userID, err := as.jwt.GetUserIDByToken(req.RefreshToken)
	if err != nil {
		as.logger.Error("failed get user id by token",
			zap.String("refresh_token", req.RefreshToken),
			zap.Error(err),
		)
		return dto.RefreshTokenResponse{}, dto.ErrGetUserIDFromToken
	}

	userRole, err := as.jwt.GetUserRoleByToken(req.RefreshToken)
	if err != nil {
		as.logger.Error("failed get user role by token",
			zap.String("refresh_token", req.RefreshToken),
			zap.Error(err),
		)
		return dto.RefreshTokenResponse{}, dto.ErrGetUserRoleFromToken
	}

	accessToken, _, err := as.jwt.GenerateToken(userID, userRole)
	if err != nil {
		as.logger.Error("failed to generate token",
			zap.String("user_id", userID),
			zap.String("user_role", userRole),
			zap.Error(err),
		)
		return dto.RefreshTokenResponse{}, dto.ErrGenerateAccessToken
	}

	as.logger.Info("refresh token success",
		zap.String("access_token", accessToken),
	)

	return dto.RefreshTokenResponse{AccessToken: accessToken}, nil
}
