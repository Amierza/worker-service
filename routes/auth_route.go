package routes

import (
	"github.com/Amierza/chat-service/handler"
	"github.com/Amierza/chat-service/jwt"
	"github.com/gin-gonic/gin"
)

func Auth(route *gin.Engine, authHandler handler.IAuthHandler, jwtService jwt.IJWT) {
	routes := route.Group("/api/v1/auth")
	{
		routes.POST("/login", authHandler.Login)
		routes.POST("/refresh-token", authHandler.RefreshToken)
	}
}
