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
	ISessionHandler interface {
		Start(ctx *gin.Context)
		Join(ctx *gin.Context)
		Leave(ctx *gin.Context)
		End(ctx *gin.Context)
		GetAll(ctx *gin.Context)
		GetDetail(ctx *gin.Context)
	}

	sessionHandler struct {
		sessionService service.ISessionService
	}
)

func NewSessionHandler(sessionService service.ISessionService) *sessionHandler {
	return &sessionHandler{
		sessionService: sessionService,
	}
}

func (sh *sessionHandler) Start(ctx *gin.Context) {
	thesisID := ctx.Param("thesis_id")
	result, err := sh.sessionService.Start(ctx, thesisID)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_START_SESSION, err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(dto.MESSAGE_SUCCESS_START_SESSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (sh *sessionHandler) Join(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")
	result, err := sh.sessionService.Join(ctx, sessionID)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_JOIN_SESSION, err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(dto.MESSAGE_SUCCESS_JOIN_SESSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (sh *sessionHandler) Leave(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")
	result, err := sh.sessionService.Leave(ctx, sessionID)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_LEAVE_SESSION, err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(dto.MESSAGE_SUCCESS_LEAVE_SESSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (sh *sessionHandler) End(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")
	result, err := sh.sessionService.End(ctx, sessionID)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_END_SESSION, err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(dto.MESSAGE_SUCCESS_END_SESSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (sh *sessionHandler) GetAll(ctx *gin.Context) {
	var (
		pagination response.PaginationRequest
		filter     dto.SessionFilterQuery
	)

	paginationParam := ctx.DefaultQuery("pagination", "true")
	usePagination := paginationParam != "false"

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_QUERY_PARAMS, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if !usePagination {
		result, err := sh.sessionService.GetAll(ctx, filter)
		if err != nil {
			status := mapErrorToStatus(err)
			res := response.BuildResponseFailed(fmt.Sprintf("%s sessions", dto.FAILED_GET_ALL), err.Error(), nil)
			ctx.AbortWithStatusJSON(status, res)
			return
		}

		res := response.BuildResponseSuccess(fmt.Sprintf("%s sessions", dto.SUCCESS_GET_ALL), result)
		ctx.JSON(http.StatusOK, res)
		return
	}

	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_QUERY_PARAMS, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := sh.sessionService.GetAllWithPagination(ctx, pagination, filter)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(fmt.Sprintf("%s sessions", dto.FAILED_GET_ALL), err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.Response{
		Status:   true,
		Messsage: fmt.Sprintf("%s sessions", dto.SUCCESS_GET_ALL),
		Data:     result.Data,
		Meta:     result.PaginationResponse,
	}
	ctx.JSON(http.StatusOK, res)
}

func (sh *sessionHandler) GetDetail(ctx *gin.Context) {
	idStr := ctx.Param("session_id")
	result, err := sh.sessionService.GetDetail(ctx, &idStr)
	if err != nil {
		status := mapErrorToStatus(err)
		res := response.BuildResponseFailed(fmt.Sprintf("%s sessions", dto.FAILED_GET_DETAIL), err.Error(), nil)
		ctx.AbortWithStatusJSON(status, res)
		return
	}

	res := response.BuildResponseSuccess(fmt.Sprintf("%s sessions", dto.SUCCESS_GET_DETAIL), result)
	ctx.JSON(http.StatusOK, res)
}
