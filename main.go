package main

import (
	"log"
	"os"
	"time"

	"github.com/Amierza/chat-service/config/database"
	"github.com/Amierza/chat-service/config/rabbitmq"
	"github.com/Amierza/chat-service/logger"
	"github.com/Amierza/chat-service/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// setup potgres connection
	db := database.SetUpPostgreSQLConnection()
	defer database.ClosePostgreSQLConnection(db)

	// setup rabbitmq connection
	rabbitConn := rabbitmq.SetUpRabbitMQConnection()
	defer rabbitmq.CloseRabbitMQConnection(rabbitConn)

	// Zap logger
	zapLogger, err := logger.New(true) // true = dev, false = prod
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer zapLogger.Sync() // flush buffer

	var (
	// JWT
	// jwt = jwt.NewJWT()

	)

	server := gin.Default()
	server.Use(middleware.CORSMiddleware())

	server.Static("/uploads", "./uploads")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	time.Local, _ = time.LoadLocation("Asia/Jakarta")

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
