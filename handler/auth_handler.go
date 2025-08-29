package handler

import (
	"net/http"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/response"
	"github.com/Amierza/chat-service/service"
	"github.com/gin-gonic/gin"
)

type (
	IAuthHandler interface {
		Login(ctx *gin.Context)
		RefreshToken(ctx *gin.Context)
	}

	authHandler struct {
		authService service.IAuthService
	}
)

func NewAuthHandler(authService service.IAuthService) *authHandler {
	return &authHandler{
		authService: authService,
	}
}

func (ah *authHandler) Login(ctx *gin.Context) {
	var payload dto.LoginRequest
	if err := ctx.ShouldBind(&payload); err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := ah.authService.Login(ctx, payload)
	if err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_LOGIN_USER, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	res := response.BuildResponseSuccess(dto.MESSAGE_SUCCESS_LOGIN_USER, result)
	ctx.JSON(http.StatusOK, res)
}

func (ah *authHandler) RefreshToken(ctx *gin.Context) {
	var payload dto.RefreshTokenRequest
	if err := ctx.ShouldBind(&payload); err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := ah.authService.RefreshToken(ctx, payload)
	if err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_REFRESH_TOKEN, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	res := response.BuildResponseSuccess(dto.MESSAGE_SUCCESS_REFRESH_TOKEN, result)
	ctx.AbortWithStatusJSON(http.StatusOK, res)
}
