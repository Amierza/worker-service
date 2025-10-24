package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Amierza/worker-service/config/database"
	"github.com/Amierza/worker-service/config/rabbitmq"
	grpcclient "github.com/Amierza/worker-service/grpc_client"
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

	// Zap logger
	zapLogger, err := logger.New(true) // true = dev, false = prod
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer zapLogger.Sync() // flush buffer

	// setup rabbitmq connection
	rabbitConn := rabbitmq.SetUpRabbitMQConnection()
	defer rabbitmq.CloseRabbitMQConnection(rabbitConn)

	// setup gRPC client ke AI Service
	grpcTarget := os.Getenv("AI_SERVICE_GRPC_ADDR")
	if grpcTarget == "" {
		grpcTarget = "localhost:50051" // default fallback
	}
	grpcClient, err := grpcclient.NewSummaryClient(grpcTarget)
	if err != nil {
		zapLogger.Fatal("failed to connect to AI gRPC service", zap.Error(err))
	}
	defer grpcClient.Close()

	var (
		// JWT
		jwt = jwt.NewJWT()

		// Consumer
		consumerRepo    = repository.NewConsumerRepository(db)
		consumerService = service.NewConsumerService(consumerRepo, zapLogger, rabbitConn, jwt, grpcClient)
		// consumerHandler = handler.NewConsumerHandler(consumerService)
	)

	// context + graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		zapLogger.Info("received shutdown signal, stopping consumer...")
		cancel()
	}()

	// jalankan consumer
	go func() {
		zapLogger.Info("starting RabbitMQ consumer listener...")
		if err := consumerService.ConsumeSummaryTasks(ctx); err != nil {
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
