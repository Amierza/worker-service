package repository

import (
	"context"

	"github.com/Amierza/worker-service/dto"
	"github.com/Amierza/worker-service/entity"
	"gorm.io/gorm"
)

type (
	IConsumerRepository interface {
		// CREATE / POST
		SaveMessages(ctx context.Context, tx *gorm.DB, task dto.TaskSummary) error

		// READ / GET

		// UPDATE / PATCH

		// DELETE / DELETE
	}

	consumerRepository struct {
		db *gorm.DB
	}
)

func NewConsumerRepository(db *gorm.DB) *consumerRepository {
	return &consumerRepository{
		db: db,
	}
}

func (cr *consumerRepository) SaveMessages(ctx context.Context, tx *gorm.DB, task dto.TaskSummary) error {
	if tx == nil {
		tx = cr.db
	}

	var messages []entity.Message
	for _, m := range task.Messages {
		messages = append(messages, entity.Message{
			ID:              m.ID,
			IsText:          m.IsText,
			Text:            m.Text,
			FileURL:         m.FileURL,
			FileType:        m.FileType,
			SenderRole:      entity.Role(m.Sender.Role),
			SenderID:        m.Sender.ID,
			SessionID:       task.SessionID,
			ParentMessageID: m.ParentMessageID,
		})
	}
	return tx.WithContext(ctx).Save(&messages).Error
}
