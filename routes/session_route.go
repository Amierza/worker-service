package routes

import (
	"github.com/Amierza/chat-service/handler"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/middleware"
	"github.com/gin-gonic/gin"
)

func Session(route *gin.Engine, sessionHandler handler.ISessionHandler, jwt jwt.IJWT) {
	routes := route.Group("/api/v1/sessions").Use(middleware.Authentication(jwt))
	{
		routes.POST("/start/:thesis_id", sessionHandler.Start)
		routes.POST("/:session_id/join", sessionHandler.Join)
		routes.POST("/:session_id/leave", sessionHandler.Leave)
		routes.POST("/:session_id/end", sessionHandler.End)
		routes.GET("", sessionHandler.GetAll)
		routes.GET("/:session_id", sessionHandler.GetDetail)
	}
}
