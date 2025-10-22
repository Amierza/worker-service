package rabbitmq

import (
	"fmt"
	"log"
	"os"

	"github.com/Amierza/worker-service/constants"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SetUpRabbitMQConnection() *amqp.Connection {
	if os.Getenv("APP_ENV") != constants.ENUM_RUN_PRODUCTION {
		if err := godotenv.Load(".env"); err != nil {
			panic(fmt.Errorf("failed to laod .env file: %v", err))
		}
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		panic("RABBITMQ_URL must be set")
	}

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	log.Println("rabbitmq connection established")
	return conn
}

func CloseRabbitMQConnection(conn *amqp.Connection) {
	err := conn.Close()
	if err != nil {
		log.Printf("error closing rabbitmq connection: %v", err)
	}
	log.Println("rabbitmq connection closed")
}
