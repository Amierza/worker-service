package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Amierza/worker-service/dto"
	"github.com/Amierza/worker-service/jwt"
	"github.com/Amierza/worker-service/repository"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type (
	IConsumerService interface {
		ConsumeSummaryTasks(ctx context.Context) error
	}

	consumerService struct {
		consumerRepo repository.IConsumerRepository
		logger       *zap.Logger
		rabbitmq     *amqp091.Connection
		jwt          jwt.IJWT
	}
)

func NewConsumerService(consumerRepo repository.IConsumerRepository, logger *zap.Logger, rabbitmq *amqp091.Connection, jwt jwt.IJWT) *consumerService {
	return &consumerService{
		consumerRepo: consumerRepo,
		logger:       logger,
		rabbitmq:     rabbitmq,
		jwt:          jwt,
	}
}

func (cs *consumerService) ConsumeSummaryTasks(ctx context.Context) error {
	ch, err := cs.rabbitmq.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	queueName := "summary_task"

	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	cs.logger.Info("✅ Worker started listening for summary tasks...")

	// Loop terus-menerus di sini (tidak di goroutine lain)
	for {
		select {
		case <-ctx.Done():
			cs.logger.Info("❌ Consumer stopped by context")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				cs.logger.Warn("RabbitMQ channel closed, reconnect needed")
				time.Sleep(5 * time.Second)
				return fmt.Errorf("channel closed")
			}

			var task dto.TaskSummary
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				cs.logger.Error("failed to unmarshal task", zap.Error(err))
				continue
			}

			cs.logger.Info("📩 Received summary task",
				zap.String("session_id", task.SessionID.String()),
				zap.Int("message_count", len(task.Messages)),
			)

			if err := cs.consumerRepo.SaveMessages(ctx, nil, task); err != nil {
				cs.logger.Error("failed to save messages to DB", zap.Error(err))
				continue
			}

			cs.logger.Info("✅ Worker finished processing task",
				zap.String("session_id", task.SessionID.String()),
			)
		}
	}
}
