package service

import (
	"github.com/Amierza/nawasena-backend/jwt"
	"github.com/Amierza/nawasena-backend/repository"
)

type (
	IUserService interface {
	}

	UserService struct {
		userRepo   repository.IUserRepository
		jwtService jwt.IJWTService
	}
)

func NewUserService(userRepo repository.IUserRepository, jwtService jwt.IJWTService) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}
