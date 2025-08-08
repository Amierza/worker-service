package service

import (
	"github.com/Amierza/go-boiler-plate/jwt"
	"github.com/Amierza/go-boiler-plate/repository"
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
