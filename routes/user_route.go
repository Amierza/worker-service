package routes

import (
	"github.com/Amierza/nawasena-backend/handler"
	"github.com/Amierza/nawasena-backend/jwt"
	"github.com/gin-gonic/gin"
)

func User(route *gin.Engine, userHandler handler.IUserHandler, jwtService jwt.IJWTService) {
	routes := route.Group("/api/v1/users")
	{
		routes.Use()
	}
}
