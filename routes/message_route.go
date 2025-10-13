package routes

import (
	"github.com/Amierza/chat-service/handler"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/middleware"
	"github.com/gin-gonic/gin"
)

func Message(route *gin.Engine, messageHandler handler.IMessageHandler, jwt jwt.IJWT) {
	routes := route.Group("/api/v1/sessions/:session_id/messages").Use(middleware.Authentication(jwt))
	{
		routes.POST("", messageHandler.Send)
		routes.GET("", messageHandler.List)
	}
}
