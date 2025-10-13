package handler

import (
	"net/http"

	"github.com/Amierza/chat-service/dto"
)

func mapErrorToStatus(err error) int {
	switch err {
	case
		// invalid input
		dto.ErrValidateToken,
		dto.ErrGetUserIDFromToken,
		dto.ErrGetThesisByID,
		dto.ErrGetActiveSessionBySessionID,
		dto.ErrSessionAlreadyStarted,
		dto.ErrUnableStartAndJoinSessionWithTheSameUser,
		dto.ErrIncorrectPassword:
		return http.StatusBadRequest
	case dto.ErrNotFound:
		return http.StatusNotFound
	case dto.ErrUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
