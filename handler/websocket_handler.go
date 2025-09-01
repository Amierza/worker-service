// handler/websocket_handler.go
package handler

import (
	"log"
	"net/http"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	upgrader websocket.Upgrader
	jwt      jwt.IJWT
}

func NewWebSocketHandler(jwt jwt.IJWT) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		jwt: jwt,
	}
}

func (h *WebSocketHandler) HandleWebSocket(ctx *gin.Context) {
	tokenString := ctx.Query("token")
	if tokenString == "" {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_TOKEN_NOT_FOUND, "missing token", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// validasi token
	_, err := h.jwt.ValidateToken(tokenString)
	if err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_TOKEN_NOT_VALID, "invalid token", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	w := ctx.Writer
	r := ctx.Request

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	userID, err := h.jwt.GetUserIDByToken(tokenString)
	if err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_TOKEN_NOT_VALID, "invalid token", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	log.Printf("User %v connected via WebSocket", userID)

	for {
		mType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received message: %s", msg)

		// Echo the message back
		err = conn.WriteMessage(mType, msg)
		if err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}
