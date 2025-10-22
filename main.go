package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Amierza/worker-service/config/database"
	"github.com/Amierza/worker-service/config/rabbitmq"
	"github.com/Amierza/worker-service/jwt"
	"github.com/Amierza/worker-service/logger"
	"github.com/Amierza/worker-service/middleware"
	"github.com/Amierza/worker-service/repository"
	"github.com/Amierza/worker-service/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		jwt = jwt.NewJWT()

		// Consumer
		consumerRepo    = repository.NewConsumerRepository(db)
		consumerService = service.NewConsumerService(consumerRepo, zapLogger, rabbitConn, jwt)
		// consumerHandler = handler.NewConsumerHandler(consumerService)
	)

	// 🧠 Jalankan consumer di background (langsung listen)
	go func() {
		zapLogger.Info("🚀 Starting RabbitMQ consumer listener...")
		if err := consumerService.ConsumeSummaryTasks(context.Background()); err != nil {
			zapLogger.Fatal("failed to start consumer", zap.Error(err))
		}
	}()

	// Optional: Gin web server (bisa tetap dipakai untuk health check)
	server := gin.Default()
	server.Use(middleware.CORSMiddleware())

	// routes.Consumer(server, consumerHandler, jwt) // opsional

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
