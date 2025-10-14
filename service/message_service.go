package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
		Event:     "new_message",
		MessageID: msgID,
		IsText:    req.IsText,
		Text:      req.Text,
		FileURL:   req.FileURL,
		FileType:  req.FileType,
		Sender: dto.CustomUserResponse{
			ID:   user.ID,
			Role: string(user.Role),
		},
		SessionID:       sID,
		ParentMessageID: req.ParentMessageID,
		Timestamp:       time.Now().String(),
	}

	if user.LecturerID != nil {
		messageEvent.Sender.Name = user.Lecturer.Name
		messageEvent.Sender.Identifier = user.Lecturer.Nip
	}

	if user.StudentID != nil {
		messageEvent.Sender.Name = user.Student.Name
		messageEvent.Sender.Identifier = user.Student.Nim
	}

	// save to Redis
	key := fmt.Sprintf("session:%s:messages", sessionID)
	data, err := json.Marshal(messageEvent)
	if err != nil {
		ms.logger.Error("failed marshal message to json", zap.Error(err))
		return dto.ErrMarshalToJSON
	}
	// save to Redis as sorted set
	score := float64(time.Now().UnixNano()) // urut berdasarkan waktu
	if err := ms.redis.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		ms.logger.Error("failed to ZADD message to redis",
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
	// tambahkan mahasiswa
	if session.Thesis.StudentID != uuid.Nil {
		receiverUserIDs = append(receiverUserIDs, session.Thesis.StudentID)
	}
	// tambahkan semua dosen pembimbing
	for _, sup := range session.Thesis.Supervisors {
		receiverUserIDs = append(receiverUserIDs, sup.LecturerID)
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
	session, found, err := ms.sessionRepo.GetActiveSessionBySessionID(ctx, nil, sessionID)
	if !found {
		ms.logger.Warn("failed get active session by session id",
			zap.String("session_id", sessionID),
		)
		return &dto.MessagePaginationResponse{}, dto.ErrNotFound
	}
	if err != nil {
		ms.logger.Error("failed get active session by session id",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.MessagePaginationResponse{}, dto.ErrGetActiveSessionBySessionID
	}

	var dataWithPaginate *dto.MessagePaginationRepositoryResponse
	switch session.Status {
	case "ongoing":
		// 2️⃣ Ambil dari Redis (chat live)
		dataWithPaginate, err = ms.messageRepo.GetAllMessageFromRedisWithPagination(ctx, nil, req, session)
	case "finished":
		// 3️⃣ Ambil dari DB (history)
		dataWithPaginate, err = ms.messageRepo.GetAllMessageWithPagination(ctx, nil, req, session)
	default:
		ms.logger.Warn("session status invalid for listing messages",
			zap.String("session_id", sessionID),
			zap.String("status", string(session.Status)),
		)
		return nil, dto.ErrInvalidSessionStatus
	}
	if err != nil {
		ms.logger.Error("failed to get messages",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return nil, dto.ErrGetAllMessageWithPagination
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
			ID:       message.ID,
			IsText:   message.IsText,
			Text:     message.Text,
			FileURL:  message.FileURL,
			FileType: message.FileType,
			Sender: dto.CustomUserResponse{
				ID:   message.Sender.ID,
				Role: string(message.Sender.Role),
			},
			SessionID:       message.SessionID,
			ParentMessageID: message.ParentMessageID,
			Timestamp:       message.TimeStamp.CreatedAt.String(),
		}

		if message.Sender.LecturerID != nil {
			data.Sender.Name = message.Sender.Lecturer.Name
			data.Sender.Identifier = message.Sender.Lecturer.Nip
		}

		if message.Sender.StudentID != nil {
			data.Sender.Name = message.Sender.Student.Name
			data.Sender.Identifier = message.Sender.Student.Nim
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
