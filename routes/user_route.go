package routes

import (
	"github.com/Amierza/go-boiler-plate/handler"
	"github.com/Amierza/go-boiler-plate/jwt"
	"github.com/gin-gonic/gin"
)

func User(route *gin.Engine, userHandler handler.IUserHandler, jwtService jwt.IJWTService) {
	routes := route.Group("/api/v1/users")
	{
		routes.Use()
	}
}
