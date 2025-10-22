package handler

import (
	"net/http"

	"github.com/Amierza/worker-service/dto"
	"github.com/Amierza/worker-service/response"
	"github.com/Amierza/worker-service/service"
	"github.com/gin-gonic/gin"
)

type (
	IConsumerHandler interface {
		StartConsumer(ctx *gin.Context)
	}

	consumerHandler struct {
		consumerService service.IConsumerService
	}
)

func NewConsumerHandler(consumerService service.IConsumerService) *consumerHandler {
	return &consumerHandler{
		consumerService: consumerService,
	}
}

func (ch *consumerHandler) StartConsumer(ctx *gin.Context) {
	err := ch.consumerService.ConsumeSummaryTasks(ctx)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(dto.FAILED_CONSUME_SUMMARY_TASKS, err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(dto.SUCCESS_CONSUME_SUMMARY_TASKS, nil)
	ctx.JSON(http.StatusOK, res)
}
