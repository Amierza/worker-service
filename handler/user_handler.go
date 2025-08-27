package handler

import (
	"github.com/Amierza/chat-service/service"
)

type (
	IUserHandler interface {
	}

	UserHandler struct {
		userService service.IUserService
	}
)

func NewUserHandler(userService service.IUserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
