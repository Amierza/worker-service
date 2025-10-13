package routes

import (
	"github.com/Amierza/chat-service/handler"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/middleware"
	"github.com/gin-gonic/gin"
)

func User(route *gin.Engine, userHandler handler.IUserHandler, jwt jwt.IJWT) {
	routes := route.Group("/api/v1/users").Use(middleware.Authentication(jwt))
	{
		routes.GET("/profile", userHandler.GetProfile)
	}
}
