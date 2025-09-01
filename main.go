package main

import (
	"log"
	"os"

	"github.com/Amierza/chat-service/cmd"
	"github.com/Amierza/chat-service/config/database"
	"github.com/Amierza/chat-service/config/rabbitmq"
	"github.com/Amierza/chat-service/config/redis"
	"github.com/Amierza/chat-service/handler"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/middleware"
	"github.com/Amierza/chat-service/repository"
	"github.com/Amierza/chat-service/routes"
	"github.com/Amierza/chat-service/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// setup potgres connection
	db := database.SetUpPostgreSQLConnection()
	defer database.ClosePostgreSQLConnection(db)

	// setup redis connection
	redisClient := redis.SetUpRedisConnection()
	defer redis.CloseRedisConnection(redisClient)

	// setup rabbitmq connection
	rabbitConn := rabbitmq.SetUpRabbitMQConnection()
	defer rabbitmq.CloseRabbitMQConnection(rabbitConn)

	if len(os.Args) > 1 {
		cmd.Command(db)
		return
	}

	var (
		// JWT
		jwt = jwt.NewJWT()

		// Authentication
		authRepo    = repository.NewAuthRepository(db)
		authService = service.NewAuthService(authRepo, jwt)
		authHandler = handler.NewAuthHandler(authService)

		// Websocket
		wsHandler = handler.NewWebSocketHandler(jwt)
	)

	server := gin.Default()
	server.Use(middleware.CORSMiddleware())

	// Websocket route
	server.GET("/ws", wsHandler.HandleWebSocket)
	// Other route
	routes.Auth(server, authHandler, jwt)

	server.Static("/assets", "./assets")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	var serve string
	if os.Getenv("APP_ENV") == "localhost" {
		serve = "127.0.0.1:" + port
	} else {
		serve = ":" + port
	}

	if err := server.Run(serve); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}
