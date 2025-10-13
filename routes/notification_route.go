package routes

import (
	"github.com/Amierza/chat-service/handler"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/middleware"
	"github.com/gin-gonic/gin"
)

func Notification(route *gin.Engine, notificationHandler handler.INotificationHandler, jwt jwt.IJWT) {
	routes := route.Group("/api/v1/notifications").Use(middleware.Authentication(jwt))
	{
		routes.GET("", notificationHandler.GetAll)
		routes.GET("/:id", notificationHandler.GetDetail)
	}
}
