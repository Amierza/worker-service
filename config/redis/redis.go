// redis/redis.go
package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Amierza/chat-service/constants"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func SetUpRedisConnection() *redis.Client {
	if os.Getenv("APP_ENV") != constants.ENUM_RUN_PRODUCTION {
		if err := godotenv.Load(".env"); err != nil {
			panic(fmt.Errorf("failed to laod .env file: %v", err))
		}
	}

	redisAddr := os.Getenv("REDIS_ADDRESS")
	redisPass := os.Getenv("REDIS_PASS")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		panic(fmt.Errorf("failed to parse REDIS_DB: %v", err))
	}
	redisProtocolStr := os.Getenv("REDIS_PROTOCOL")
	redisProtocol, err := strconv.Atoi(redisProtocolStr)
	if err != nil {
		panic(fmt.Errorf("failed to parse REDIS_PROTOCOL: %v", err))
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
		Protocol: redisProtocol,
	})

	// Cek koneksi
	_, err = client.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf("failed to connect redis: %v", err))
	}

	log.Println("redis connection established")
	return client
}

func CloseRedisConnection(client *redis.Client) {
	err := client.Close()
	if err != nil {
		log.Printf("Error closing redis connection: %v", err)
	}
	log.Println("redis connection closed")
}
