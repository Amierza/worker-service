package handler

import (
	"net/http"

	"github.com/Amierza/worker-service/dto"
)

func mapErrorToStatus(err error) int {
	switch err {
	case
		// invalid input
		dto.ErrValidateToken,
		dto.ErrGetUserIDFromToken:
		return http.StatusBadRequest
	case dto.ErrNotFound:
		return http.StatusNotFound
	case dto.ErrUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
