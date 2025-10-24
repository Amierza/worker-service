package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/Amierza/ai-service/proto"
	"github.com/Amierza/worker-service/dto"
	grpcclient "github.com/Amierza/worker-service/grpc_client"
	"github.com/Amierza/worker-service/jwt"
	"github.com/Amierza/worker-service/repository"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type (
	IConsumerService interface {
		ConsumeSummaryTasks(ctx context.Context) error
	}

	consumerService struct {
		consumerRepo repository.IConsumerRepository
		logger       *zap.Logger
		rabbitmq     *amqp.Connection
		jwt          jwt.IJWT
		grpcClient   *grpcclient.SummaryClient
	}
)

func NewConsumerService(consumerRepo repository.IConsumerRepository, logger *zap.Logger, rabbitmq *amqp.Connection, jwt jwt.IJWT, grpcClient *grpcclient.SummaryClient) *consumerService {
	return &consumerService{
		consumerRepo: consumerRepo,
		logger:       logger,
		rabbitmq:     rabbitmq,
		jwt:          jwt,
		grpcClient:   grpcClient,
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
			cs.logger.Info("consumer stopped by context")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				cs.logger.Warn("rabbitMQ channel closed, reconnect needed")
				time.Sleep(5 * time.Second)
				return fmt.Errorf("channel closed")
			}

			var task dto.TaskSummary
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				cs.logger.Error("failed to unmarshal task", zap.Error(err))
				continue
			}

			cs.logger.Info("received summary task",
				zap.String("session_id", task.SessionID.String()),
				zap.Int("message_count", len(task.Messages)),
			)

			// panggil gRPC ke AI service untuk membuat ringkasan
			req := &pb.SummaryRequest{
				Task: &pb.TaskSummary{
					SessionId:     task.SessionID.String(),
					SessionStatus: task.SessionStatus,
					StartedAt:     task.StartedAt.String(),
					EndedAt:       task.EndedAt.String(),
					CreatedAt:     task.CreatedAt.String(),
					Owner: &pb.CustomUser{
						Id:         task.Owner.ID.String(),
						Name:       task.Owner.Name,
						Identifier: task.Owner.Identifier,
						Role:       task.Owner.Role,
					},
					Student: &pb.Student{
						Id:    task.Student.ID.String(),
						Nim:   task.Student.Nim,
						Name:  task.Student.Name,
						Email: task.Student.Email,
						StudyProgram: &pb.StudyProgram{
							Id:     task.Student.StudyProgram.ID.String(),
							Name:   task.Student.StudyProgram.Name,
							Degree: string(task.Student.StudyProgram.Degree),
							Faculty: &pb.Faculty{
								Id:   task.Student.StudyProgram.Faculty.ID.String(),
								Name: task.Student.StudyProgram.Faculty.Name,
							},
						},
					},
					Supervisors: func() []*pb.Lecturer {
						var supervisors []*pb.Lecturer
						for _, sup := range task.Supervisors {
							supervisors = append(supervisors, &pb.Lecturer{
								Id:           sup.ID.String(),
								Nip:          sup.Nip,
								Name:         sup.Name,
								Email:        sup.Email,
								TotalStudent: int32(sup.TotalStudent),
								StudyProgram: &pb.StudyProgram{
									Id:     sup.StudyProgram.ID.String(),
									Name:   sup.StudyProgram.Name,
									Degree: string(sup.StudyProgram.Degree),
									Faculty: &pb.Faculty{
										Id:   sup.StudyProgram.Faculty.ID.String(),
										Name: sup.StudyProgram.Faculty.Name,
									},
								},
							})
						}
						return supervisors
					}(),
					ThesisInfo: &pb.ThesisInfo{
						Title:       task.ThesisInfo.Title,
						Progress:    string(task.ThesisInfo.Progress),
						Description: task.ThesisInfo.Description,
					},
					Messages: func() []*pb.MessageSummary {
						var messages []*pb.MessageSummary
						for _, msg := range task.Messages {
							messages = append(messages, &pb.MessageSummary{
								Id:       msg.ID.String(),
								IsText:   msg.IsText,
								Text:     msg.Text,
								FileUrl:  msg.FileURL,
								FileType: msg.FileType,
								Sender: &pb.CustomUser{
									Id:         msg.Sender.ID.String(),
									Name:       msg.Sender.Name,
									Identifier: msg.Sender.Identifier,
									Role:       msg.Sender.Role,
								},
								ParentMessageId: func() string {
									if msg.ParentMessageID != nil {
										return msg.ParentMessageID.String()
									}
									return ""
								}(),
								Timestamp: msg.Timestamp,
							})
						}
						return messages
					}(),
				},
			}

			_, err := cs.grpcClient.GenerateSummary(ctx, req)
			if err != nil {
				cs.logger.Error("failed to generate summary via gRPC", zap.Error(err))
				continue
			}

			cs.logger.Info("summary successfully generated",
				zap.String("session_id", task.SessionID.String()),
			)

			if err := cs.consumerRepo.SaveMessages(ctx, nil, task); err != nil {
				cs.logger.Error("failed to save messages to DB", zap.Error(err))
				continue
			}

			cs.logger.Info("worker finished processing task",
				zap.String("session_id", task.SessionID.String()),
			)
		}
	}
}
