package repository

import (
	"context"
	"errors"

	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

type (
	INotificationRepository interface {
		// CREATE / POST
		CreateNotification(ctx context.Context, tx *gorm.DB, notification *entity.Notification) error

		// READ / GET
		GetAllNotificationsByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]*entity.Notification, error)
		GetNotificationByID(ctx context.Context, tx *gorm.DB, id *string) (*entity.Notification, bool, error)

		// UPDATE / PATCH
		UpdateIsRead(ctx context.Context, tx *gorm.DB, id *string) error

		// DELETE / DELETE
	}

	notificationRepository struct {
		db *gorm.DB
	}
)

func NewNotificationRepository(db *gorm.DB) *notificationRepository {
	return &notificationRepository{
		db: db,
	}
}

// CREATE / POST
func (nr *notificationRepository) CreateNotification(ctx context.Context, tx *gorm.DB, notification *entity.Notification) error {
	if tx == nil {
		tx = nr.db
	}

	return tx.WithContext(ctx).Create(&notification).Error
}

// READ / GET
func (nr *notificationRepository) GetAllNotificationsByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]*entity.Notification, error) {
	if tx == nil {
		tx = nr.db
	}

	var (
		notifications []*entity.Notification
		err           error
	)

	query := tx.WithContext(ctx).
		Preload("User.SessionOwners.Thesis").
		Preload("User.Student.StudyProgram.Faculty").
		Preload("User.Lecturer.StudyProgram.Faculty").
		Where("user_id = ?", userID).
		Model(&entity.Notification{})
	if err := query.Order(`"created_at" DESC`).Find(&notifications).Error; err != nil {
		return []*entity.Notification{}, err
	}

	return notifications, err
}
func (nr *notificationRepository) GetNotificationByID(ctx context.Context, tx *gorm.DB, id *string) (*entity.Notification, bool, error) {
	if tx == nil {
		tx = nr.db
	}

	var notification *entity.Notification
	err := tx.WithContext(ctx).
		Preload("User.SessionOwners.Thesis").
		Preload("User.Student.StudyProgram.Faculty").
		Preload("User.Lecturer.StudyProgram.Faculty").
		Where("id = ?", id).
		Take(&notification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.Notification{}, false, nil
	}
	if err != nil {
		return &entity.Notification{}, false, err
	}

	return notification, true, nil
}

// UPDATE / PATCH
func (nr *notificationRepository) UpdateIsRead(ctx context.Context, tx *gorm.DB, id *string) error {
	if tx == nil {
		tx = nr.db
	}

	err := tx.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Update("is_read", true)

	if err.Error != nil {
		return err.Error
	}

	if err.RowsAffected == 0 {
		return errors.New("notification not found or no change made")
	}

	return nil
}
