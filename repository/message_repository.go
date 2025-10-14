package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/Amierza/chat-service/constants"
	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/entity"
	"github.com/Amierza/chat-service/response"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	IMessageRepository interface {
		// CREATE / POST
		CreateMessage(ctx context.Context, tx *gorm.DB, message *entity.Message) error

		// READ / GET
		GetAllMessageFromRedisWithPagination(ctx context.Context, tx *gorm.DB, req response.PaginationRequest, session *entity.Session) (*dto.MessagePaginationRepositoryResponse, error)
		GetAllMessageWithPagination(ctx context.Context, tx *gorm.DB, req response.PaginationRequest, session *entity.Session) (*dto.MessagePaginationRepositoryResponse, error)

		// UPDATE / PATCH

		// DELETE / DELETE
	}

	messageRepository struct {
		db     *gorm.DB
		logger *zap.Logger
		redis  *redis.Client
	}
)

func NewMessageRepository(db *gorm.DB, logger *zap.Logger, redis *redis.Client) *messageRepository {
	return &messageRepository{
		db:     db,
		logger: logger,
		redis:  redis,
	}
}

// CREATE / POST
func (mr *messageRepository) CreateMessage(ctx context.Context, tx *gorm.DB, message *entity.Message) error {
	if tx == nil {
		tx = mr.db
	}

	return tx.WithContext(ctx).Create(&message).Error
}

// READ / GET
func (mr *messageRepository) GetAllMessageFromRedisWithPagination(ctx context.Context, tx *gorm.DB, req response.PaginationRequest, session *entity.Session) (*dto.MessagePaginationRepositoryResponse, error) {
	if req.PerPage == 0 {
		req.PerPage = 10
	}
	if req.Page == 0 {
		req.Page = 1
	}

	start := int64((req.Page - 1) * req.PerPage)
	end := start + int64(req.PerPage) - 1

	key := fmt.Sprintf("session:%s:messages", session.ID)

	// Get total count
	count, err := mr.redis.ZCard(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get message count from redis: %w", err)
	}

	if count == 0 {
		return &dto.MessagePaginationRepositoryResponse{
			Messages: []entity.Message{},
			PaginationResponse: response.PaginationResponse{
				Page:    req.Page,
				PerPage: req.PerPage,
				MaxPage: 0,
				Count:   0,
			},
		}, nil
	}

	// Ambil data dari Redis (terbaru ke terlama)
	results, err := mr.redis.ZRevRange(ctx, key, start, end).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get messages from redis: %w", err)
	}

	var messages []entity.Message
	for _, raw := range results {
		var evt dto.MessageEventPublish
		if err := json.Unmarshal([]byte(raw), &evt); err != nil {
			mr.logger.Warn("failed to unmarshal redis message", zap.Error(err))
			continue
		}

		msg := entity.Message{
			ID:       evt.MessageID,
			IsText:   evt.IsText,
			Text:     evt.Text,
			FileURL:  evt.FileURL,
			FileType: evt.FileType,
			Sender: entity.User{
				ID:         evt.Sender.ID,
				Identifier: evt.Sender.Identifier,
				Role:       entity.Role(evt.Sender.Role),
			},
			SessionID:       evt.SessionID,
			ParentMessageID: evt.ParentMessageID,
		}

		if evt.Sender.Role == constants.ENUM_ROLE_STUDENT {
			msg.Sender.StudentID = &session.Thesis.Student.ID
		}

		if evt.Sender.Role == constants.ENUM_ROLE_LECTURER {
			for _, supervisor := range session.Thesis.Supervisors {
				if evt.Sender.Identifier == supervisor.Lecturer.Nip {
					msg.Sender.LecturerID = &supervisor.Lecturer.ID
				}
			}
		}

		messages = append(messages, msg)
	}

	totalPage := int64(math.Ceil(float64(count) / float64(req.PerPage)))

	return &dto.MessagePaginationRepositoryResponse{
		Messages: messages,
		PaginationResponse: response.PaginationResponse{
			Page:    req.Page,
			PerPage: req.PerPage,
			MaxPage: totalPage,
			Count:   count,
		},
	}, nil
}

func (mr *messageRepository) GetAllMessageWithPagination(ctx context.Context, tx *gorm.DB, req response.PaginationRequest, session *entity.Session) (*dto.MessagePaginationRepositoryResponse, error) {
	if tx == nil {
		tx = mr.db
	}

	var messages []entity.Message
	var err error
	var count int64

	if req.PerPage == 0 {
		req.PerPage = 10
	}

	if req.Page == 0 {
		req.Page = 1
	}

	query := tx.WithContext(ctx).
		Model(&entity.Message{}).
		Preload("Sender.Student.StudyProgram.Faculty").
		Preload("Sender.Lecturer.StudyProgram.Faculty").
		Where("session_id = ?", session.ID)

	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	if err := query.Order(`"created_at" DESC`).Scopes(response.Paginate(req.Page, req.PerPage)).Find(&messages).Error; err != nil {
		return nil, err
	}

	totalPage := int64(math.Ceil(float64(count) / float64(req.PerPage)))

	return &dto.MessagePaginationRepositoryResponse{
		Messages: messages,
		PaginationResponse: response.PaginationResponse{
			Page:    req.Page,
			PerPage: req.PerPage,
			MaxPage: totalPage,
			Count:   count,
		},
	}, err
}

// UPDATE / PATCH

// DELETE / DELETE
