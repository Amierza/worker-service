package service

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/helper"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type (
	IWebsocketService interface {
		HandleWebSocket(ctx *gin.Context)
		SendToUser(userID string, message []byte) error
	}

	webSocketService struct {
		upgrader    websocket.Upgrader
		jwt         jwt.IJWT
		redis       *redis.Client
		connections map[string]*websocket.Conn
		mu          sync.RWMutex
	}
)

func NewWebSocketService(jwt jwt.IJWT, redis *redis.Client) *webSocketService {
	return &webSocketService{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		jwt:         jwt,
		redis:       redis,
		connections: make(map[string]*websocket.Conn),
	}
}

func (wh *webSocketService) HandleWebSocket(ctx *gin.Context) {
	tokenString := ctx.Query("token")
	if tokenString == "" {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_TOKEN_NOT_FOUND, "missing token", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// validasi token
	_, err := wh.jwt.ValidateToken(tokenString)
	if err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_TOKEN_NOT_VALID, "invalid token", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	userID, err := wh.jwt.GetUserIDByToken(tokenString)
	if err != nil {
		res := response.BuildResponseFailed(dto.MESSAGE_FAILED_TOKEN_NOT_VALID, "invalid token", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	conn, err := wh.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer func() {
		wh.mu.Lock()
		delete(wh.connections, userID)
		wh.mu.Unlock()

		helper.SetOffline(userID)
		conn.Close()
	}()

	wh.mu.Lock()
	wh.connections[userID] = conn
	wh.mu.Unlock()

	helper.SetOnline(userID)
	log.Printf("User %v connected via WebSocket", userID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		log.Printf("Received message from %s: %s", userID, msg)
	}
}

// SendToUser mengirim message langsung ke user tertentu
func (wh *webSocketService) SendToUser(userID string, message []byte) error {
	wh.mu.RLock()
	conn, ok := wh.connections[userID]
	wh.mu.RUnlock()

	if !ok {
		// user tidak online → biar Start() bisa fallback bikin notif
		return fmt.Errorf("user %s not connected", userID)
	}

	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("Error sending message to %s: %v", userID, err)

		// kalau koneksi rusak → remove biar gak nyangkut
		wh.mu.Lock()
		delete(wh.connections, userID)
		wh.mu.Unlock()

		return err
	}

	log.Printf("[WS SEND] to userID: '%s'", userID)
	return nil
}
