package service

import (
	"context"
	"log"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/repository"
	"go.uber.org/zap"
)

type (
	INotificationService interface {
		GetAll(ctx context.Context) ([]*dto.NotificationResponse, error)
		GetDetail(ctx context.Context, id *string) (*dto.NotificationResponse, error)
	}

	notificationService struct {
		notificationRepo repository.INotificationRepository
		logger           *zap.Logger
		jwt              jwt.IJWT
	}
)

func NewNotificationService(notificationRepo repository.INotificationRepository, logger *zap.Logger, jwt jwt.IJWT) *notificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		logger:           logger,
		jwt:              jwt,
	}
}

func (ns *notificationService) GetAll(ctx context.Context) ([]*dto.NotificationResponse, error) {
	token := ctx.Value("Authorization").(string)
	userIDString, err := ns.jwt.GetUserIDByToken(token)
	if err != nil {
		return nil, dto.ErrGetUserIDFromToken
	}
	log.Println(userIDString)

	datas, err := ns.notificationRepo.GetAllNotificationsByUserID(ctx, nil, userIDString)
	if err != nil {
		ns.logger.Error("failed to get all notifications",
			zap.Error(err),
		)
		return nil, dto.ErrGetAllNotificationsByUserID
	}

	notifications := make([]*dto.NotificationResponse, 0, len(datas))
	for _, data := range datas {
		notifications = append(notifications, &dto.NotificationResponse{
			ID:      data.ID,
			UserID:  data.UserID,
			Title:   data.Title,
			Message: data.Message,
			IsRead:  data.IsRead,
		})
	}
	ns.logger.Info("success get all notifications",
		zap.Int("count", len(datas)),
	)

	return notifications, nil
}

func (ns *notificationService) GetDetail(ctx context.Context, id *string) (*dto.NotificationResponse, error) {
	data, found, err := ns.notificationRepo.GetNotificationByID(ctx, nil, id)
	if err != nil {
		ns.logger.Error("failed to get notification by id",
			zap.String("id", *id),
			zap.Error(err),
		)
		return nil, dto.ErrGetNotificationByID
	}
	if !found {
		ns.logger.Warn("notification not found",
			zap.String("id", *id),
		)
		return nil, dto.ErrNotFound
	}

	err = ns.notificationRepo.UpdateIsRead(ctx, nil, id)
	if err != nil {
		ns.logger.Error("failed to update is_read notification",
			zap.String("id", *id),
			zap.Bool("is_read", data.IsRead),
			zap.Error(err),
		)
		return nil, dto.ErrUpdateIsReadNotification
	}
	data.IsRead = true
	ns.logger.Info("success update is_read notification",
		zap.String("id", *id),
	)

	notification := &dto.NotificationResponse{
		ID:      data.ID,
		UserID:  data.UserID,
		Title:   data.Title,
		Message: data.Message,
		IsRead:  data.IsRead,
	}
	ns.logger.Info("success get detail notification",
		zap.String("id", *id),
	)

	return notification, nil
}
