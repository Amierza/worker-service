package handler

import (
	"fmt"
	"net/http"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/response"
	"github.com/Amierza/chat-service/service"
	"github.com/gin-gonic/gin"
)

type (
	IUserHandler interface {
		GetProfile(ctx *gin.Context)
	}

	userHandler struct {
		userService service.IUserService
	}
)

func NewUserHandler(userService service.IUserService) *userHandler {
	return &userHandler{
		userService: userService,
	}
}

func (uh *userHandler) GetProfile(ctx *gin.Context) {
	result, err := uh.userService.GetProfile(ctx)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(fmt.Sprintf("%s user", dto.FAILED_GET_PROFILE), err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(fmt.Sprintf("%s user", dto.SUCCESS_GET_PROFILE), result)
	ctx.JSON(http.StatusOK, res)
}
