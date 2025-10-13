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
	IMessageHandler interface {
		Send(ctx *gin.Context)
		List(ctx *gin.Context)
	}

	messageHandler struct {
		messageService service.IMessageService
	}
)

func NewMessageHandler(messageService service.IMessageService) *messageHandler {
	return &messageHandler{
		messageService: messageService,
	}
}

func (mh *messageHandler) Send(ctx *gin.Context) {
	var payload dto.SendMessageRequest
	if err := ctx.ShouldBind(&payload); err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_SEND_MESSAGE, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	sessionID := ctx.Param("session_id")
	err := mh.messageService.Send(ctx, payload, sessionID)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_SEND_MESSAGE, err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(dto.MESSAGE_SUCCESS_SEND_MESSAGE, nil)
	ctx.JSON(http.StatusCreated, res)
}

func (mh *messageHandler) List(ctx *gin.Context) {
	var payload response.PaginationRequest
	if err := ctx.ShouldBind(&payload); err != nil {
		res := response.BuildResponseFailed(fmt.Sprintf("%s messages", dto.FAILED_GET_ALL), err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	sessionID := ctx.Param("session_id")
	result, err := mh.messageService.List(ctx, payload, sessionID)
	if err != nil {
		res := response.BuildResponseFailed(fmt.Sprintf("%s messages", dto.FAILED_GET_ALL), err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response{
		Status:   true,
		Messsage: fmt.Sprintf("%s messages", dto.SUCCESS_GET_ALL),
		Data:     result.Data,
		Meta:     result.PaginationResponse,
	}

	ctx.JSON(http.StatusOK, res)
}
