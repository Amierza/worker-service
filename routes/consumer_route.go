package routes

import (
	"github.com/Amierza/worker-service/handler"
	"github.com/Amierza/worker-service/jwt"
	"github.com/Amierza/worker-service/middleware"
	"github.com/gin-gonic/gin"
)

func Consumer(route *gin.Engine, consumerHandler handler.IConsumerHandler, jwt jwt.IJWT) {
	routes := route.Group("/api/v1/consumers").Use(middleware.Authentication(jwt))
	{
		routes.GET("/start", consumerHandler.StartConsumer)
	}
}
