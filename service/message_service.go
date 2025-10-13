package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Amierza/chat-service/constants"
	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/repository"
	"github.com/Amierza/chat-service/response"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type (
	IMessageService interface {
		Send(ctx context.Context, req dto.SendMessageRequest, sessionID string) error
		List(ctx context.Context, req response.PaginationRequest, sessionID string) (*dto.MessagePaginationResponse, error)
	}

	messageService struct {
		messageRepo repository.IMessageRepository
		sessionRepo repository.ISessionRepository
		userRepo    repository.IUserRepository
		logger      *zap.Logger
		wsService   IWebsocketService
		jwt         jwt.IJWT
		redis       *redis.Client
	}
)

func NewMessageService(messageRepo repository.IMessageRepository, sessionRepo repository.ISessionRepository, userRepo repository.IUserRepository, logger *zap.Logger, wsService IWebsocketService, jwt jwt.IJWT, redis *redis.Client) *messageService {
	return &messageService{
		messageRepo: messageRepo,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		logger:      logger,
		wsService:   wsService,
		jwt:         jwt,
		redis:       redis,
	}
}

func (ms *messageService) Send(ctx context.Context, req dto.SendMessageRequest, sessionID string) error {
	// get information user login
	token := ctx.Value("Authorization").(string)
	userIDString, err := ms.jwt.GetUserIDByToken(token)
	if err != nil {
		ms.logger.Error("failed to extract user_id from token",
			zap.String("access_token", token),
			zap.Error(err),
		)
		return dto.ErrGetUserIDFromToken
	}
	user, found, err := ms.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		ms.logger.Error("failed to fetch user by id",
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return dto.ErrGetUserByID
	}
	if !found {
		ms.logger.Warn("user not found",
			zap.String("user_id", userIDString),
		)
		return dto.ErrNotFound
	}

	// validate active session
	session, found, _ := ms.sessionRepo.GetActiveSessionBySessionID(ctx, nil, sessionID)
	if !found {
		ms.logger.Warn("failed get active session by session id",
			zap.String("session_id", sessionID),
		)
		return dto.ErrNotFound
	}

	// cannot send if session is not ongoing
	if session.Status == "waiting" {
		ms.logger.Error("failed to send message because session has not started yet",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return dto.ErrSessionWaiting
	}
	if session.Status == "finished" {
		ms.logger.Error("failed to send message because session is finished",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return dto.ErrSessionFinished
	}

	// parse session id
	sID, err := uuid.Parse(sessionID)
	if err != nil {
		ms.logger.Error("failed parse session id to uuid",
			zap.String("session_id", sessionID),
		)
		return dto.ErrParseStringToUUID
	}

	// create message event
	msgID := uuid.New()
	messageEvent := &dto.MessageEventPublish{
		Event:           "new_message",
		MessageID:       msgID,
		IsText:          req.IsText,
		Text:            req.Text,
		FileURL:         req.FileURL,
		FileType:        req.FileType,
		SenderRole:      user.Role,
		SenderID:        user.ID,
		SessionID:       sID,
		ParentMessageID: req.ParentMessageID,
	}

	// save to Redis
	key := fmt.Sprintf("session:%s:messages", sessionID)
	data, err := json.Marshal(messageEvent)
	if err != nil {
		ms.logger.Error("failed marshal message to json", zap.Error(err))
		return dto.ErrMarshalToJSON
	}
	if err := ms.redis.RPush(ctx, key, data).Err(); err != nil {
		ms.logger.Error("failed push message to redis",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return dto.ErrPushToRedis
	}

	// set TTL expired
	if err := ms.redis.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
		ms.logger.Warn("failed to set ttl for redis key",
			zap.String("key", key),
			zap.Error(err),
		)
	}
	ms.logger.Info("success to save message to redis",
		zap.String("message_id", msgID.String()),
		zap.String("session_id", sessionID),
	)

	var receiverUserIDs []uuid.UUID
	if user.StudentID != nil && session.Thesis.StudentID != uuid.Nil && *user.StudentID == session.Thesis.StudentID {
		for _, sup := range session.Thesis.Supervisors {
			receiverUserIDs = append(receiverUserIDs, sup.LecturerID)
		}
	} else if user.Role == constants.ENUM_ROLE_LECTURER {
		if session.Thesis.StudentID != uuid.Nil {
			receiverUserIDs = append(receiverUserIDs, session.Thesis.StudentID)
		}
		for _, sup := range session.Thesis.Supervisors {
			if sup.LecturerID != user.Lecturer.ID {
				receiverUserIDs = append(receiverUserIDs, sup.LecturerID)
			}
		}
	} else {
		return dto.ErrUnauthorized
	}

	// send message to all receiver
	for _, receiverID := range receiverUserIDs {
		receiverUser, found, err := ms.userRepo.GetUserByStudentOrLecturerID(ctx, nil, receiverID.String())
		if err != nil {
			ms.logger.Error("failed to resolve receiver user",
				zap.String("receiver_entity_id", receiverID.String()),
				zap.Error(err),
			)
			continue
		}
		if !found {
			ms.logger.Warn("receiver user not found for entity_id",
				zap.String("receiver_entity_id", receiverID.String()),
			)
			continue
		}

		dataEvent, _ := json.Marshal(messageEvent)
		if err := ms.wsService.SendToUser(receiverUser.ID.String(), dataEvent); err != nil {
			ms.logger.Error("failed to send websocket message",
				zap.String("receiver_user_id", receiverUser.ID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (ms *messageService) List(ctx context.Context, req response.PaginationRequest, sessionID string) (*dto.MessagePaginationResponse, error) {
	// validate active session
	_, found, _ := ms.sessionRepo.GetActiveSessionBySessionID(ctx, nil, sessionID)
	if !found {
		ms.logger.Warn("failed get active session by session id",
			zap.String("session_id", sessionID),
		)
		return &dto.MessagePaginationResponse{}, dto.ErrNotFound
	}

	// get all messages
	dataWithPaginate, err := ms.messageRepo.GetAllMessageWithPagination(ctx, nil, req, sessionID)
	if err != nil {
		ms.logger.Error("failed to get all messages with pagination",
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
			zap.Error(err),
		)
		return &dto.MessagePaginationResponse{}, dto.ErrGetAllMessageWithPagination
	}
	ms.logger.Info("success get all messages with pagination",
		zap.Int("page", dataWithPaginate.Page),
		zap.Int("per_page", dataWithPaginate.PerPage),
		zap.Int64("count", dataWithPaginate.Count),
	)

	// loop for build responses
	var datas []dto.MessageResponse
	for _, message := range dataWithPaginate.Messages {
		data := dto.MessageResponse{
			ID:              message.ID,
			IsText:          message.IsText,
			Text:            message.Text,
			FileURL:         message.FileURL,
			FileType:        message.FileType,
			SenderID:        message.SenderID,
			SessionID:       message.SessionID,
			ParentMessageID: message.ParentMessageID,
		}

		datas = append(datas, data)
	}

	return &dto.MessagePaginationResponse{
		Data: datas,
		PaginationResponse: response.PaginationResponse{
			Page:    dataWithPaginate.Page,
			PerPage: dataWithPaginate.PerPage,
			MaxPage: dataWithPaginate.MaxPage,
			Count:   dataWithPaginate.Count,
		},
	}, nil
}
