package repository

import (
	"context"
	"math"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/entity"
	"github.com/Amierza/chat-service/response"
	"gorm.io/gorm"
)

type (
	IMessageRepository interface {
		// CREATE / POST
		CreateMessage(ctx context.Context, tx *gorm.DB, message *entity.Message) error

		// READ / GET
		GetAllMessageWithPagination(ctx context.Context, tx *gorm.DB, req response.PaginationRequest, sessionID string) (dto.MessagePaginationRepositoryResponse, error)

		// UPDATE / PATCH

		// DELETE / DELETE
	}

	messageRepository struct {
		db *gorm.DB
	}
)

func NewMessageRepository(db *gorm.DB) *messageRepository {
	return &messageRepository{
		db: db,
	}
}

// CREATE / POST
func (nr *messageRepository) CreateMessage(ctx context.Context, tx *gorm.DB, message *entity.Message) error {
	if tx == nil {
		tx = nr.db
	}

	return tx.WithContext(ctx).Create(&message).Error
}

// READ / GET
func (cdr *messageRepository) GetAllMessageWithPagination(ctx context.Context, tx *gorm.DB, req response.PaginationRequest, sessionID string) (dto.MessagePaginationRepositoryResponse, error) {
	if tx == nil {
		tx = cdr.db
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
		Preload("User").
		Where("session_id = ?", sessionID)

	if err := query.Count(&count).Error; err != nil {
		return dto.MessagePaginationRepositoryResponse{}, err
	}

	if err := query.Order(`"created_at" DESC`).Scopes(response.Paginate(req.Page, req.PerPage)).Find(&messages).Error; err != nil {
		return dto.MessagePaginationRepositoryResponse{}, err
	}

	totalPage := int64(math.Ceil(float64(count) / float64(req.PerPage)))

	return dto.MessagePaginationRepositoryResponse{
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
