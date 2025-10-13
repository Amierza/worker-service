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
	INotificationHandler interface {
		GetAll(ctx *gin.Context)
		GetDetail(ctx *gin.Context)
	}

	notificationHandler struct {
		notificationService service.INotificationService
	}
)

func NewNotificationHandler(notificationService service.INotificationService) *notificationHandler {
	return &notificationHandler{
		notificationService: notificationService,
	}
}

func (nh *notificationHandler) GetAll(ctx *gin.Context) {
	result, err := nh.notificationService.GetAll(ctx)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(fmt.Sprintf("%s notifications", dto.FAILED_GET_ALL), err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(fmt.Sprintf("%s notifications", dto.SUCCESS_GET_ALL), result)
	ctx.JSON(http.StatusOK, res)
}

func (nh *notificationHandler) GetDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	result, err := nh.notificationService.GetDetail(ctx, &idStr)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(fmt.Sprintf("%s notifications", dto.FAILED_GET_DETAIL), err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(fmt.Sprintf("%s notifications", dto.SUCCESS_GET_DETAIL), result)
	ctx.JSON(http.StatusOK, res)
}
