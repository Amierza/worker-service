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
	"github.com/Amierza/chat-service/logger"
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

	// Zap logger
	zapLogger, err := logger.New(true) // true = dev, false = prod
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer zapLogger.Sync() // flush buffer

	var (
		// JWT
		jwt = jwt.NewJWT()

		// Authentication
		authRepo    = repository.NewAuthRepository(db)
		authService = service.NewAuthService(authRepo, zapLogger, jwt)
		authHandler = handler.NewAuthHandler(authService)

		// Websocket
		wsService = service.NewWebSocketService(jwt, redisClient)

		// User
		userRepo    = repository.NewUserRepository(db)
		userService = service.NewUserService(userRepo, zapLogger, jwt)
		userHandler = handler.NewUserHandler(userService)

		// Notification
		notificationRepo    = repository.NewNotificationRepository(db)
		notificationService = service.NewNotificationService(notificationRepo, zapLogger, jwt)
		notificationHandler = handler.NewNotificationHandler(notificationService)

		// Session
		sessionRepo    = repository.NewSessionRepository(db)
		sessionService = service.NewSessionService(sessionRepo, notificationRepo, userRepo, zapLogger, wsService, jwt, redisClient)
		sessionHandler = handler.NewSessionHandler(sessionService)

		// Message
		messageRepo    = repository.NewMessageRepository(db)
		messageService = service.NewMessageService(messageRepo, sessionRepo, userRepo, zapLogger, wsService, jwt, redisClient)
		messageHandler = handler.NewMessageHandler(messageService)
	)

	server := gin.Default()
	server.Use(middleware.CORSMiddleware())

	// Websocket route
	server.GET("/ws", wsService.HandleWebSocket)
	// Other route
	routes.Auth(server, authHandler, jwt)
	routes.User(server, userHandler, jwt)
	routes.Notification(server, notificationHandler, jwt)
	routes.Session(server, sessionHandler, jwt)
	routes.Message(server, messageHandler, jwt)

	server.Static("/uploads", "./uploads")

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
