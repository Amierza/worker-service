package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Amierza/chat-service/constants"
	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/entity"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/repository"
	"github.com/Amierza/chat-service/response"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type (
	ISessionService interface {
		Start(ctx context.Context, thesisID string) (*dto.SessionResponse, error)
		Join(ctx context.Context, sessionID string) (*dto.SessionResponse, error)
		Leave(ctx context.Context, sessionID string) (*dto.SessionResponse, error)
		End(ctx context.Context, sessionID string) (*dto.SessionResponse, error)
		GetAll(ctx context.Context, filter dto.SessionFilterQuery) ([]*dto.SessionResponse, error)
		GetAllWithPagination(ctx context.Context, req response.PaginationRequest, filter dto.SessionFilterQuery) (dto.SessionPaginationResponse, error)
		GetDetail(ctx context.Context, id *string) (*dto.SessionResponse, error)
	}

	sessionService struct {
		sessionRepo      repository.ISessionRepository
		notificationRepo repository.INotificationRepository
		userRepo         repository.IUserRepository
		logger           *zap.Logger
		wsService        IWebsocketService
		jwt              jwt.IJWT
		redis            *redis.Client
	}
)

func NewSessionService(sessionRepo repository.ISessionRepository, notificationRepo repository.INotificationRepository, userRepo repository.IUserRepository, logger *zap.Logger, wsService IWebsocketService, jwt jwt.IJWT, redis *redis.Client) *sessionService {
	return &sessionService{
		sessionRepo:      sessionRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		logger:           logger,
		wsService:        wsService,
		jwt:              jwt,
		redis:            redis,
	}
}

func mapStudent(u entity.User) *dto.StudentResponse {
	if u.StudentID == nil {
		return nil
	}
	return &dto.StudentResponse{
		ID:    u.Student.ID,
		Nim:   u.Student.Nim,
		Name:  u.Student.Name,
		Email: u.Student.Email,
		StudyProgram: dto.StudyProgramResponse{
			ID:     u.Student.StudyProgramID,
			Name:   u.Student.StudyProgram.Name,
			Degree: u.Student.StudyProgram.Degree,
			Faculty: dto.FacultyResponse{
				ID:   u.Student.StudyProgram.FacultyID,
				Name: u.Student.StudyProgram.Faculty.Name,
			},
		},
	}
}

func mapLecturer(u entity.User) *dto.LecturerResponse {
	if u.LecturerID == nil {
		return nil
	}
	return &dto.LecturerResponse{
		ID:           u.Lecturer.ID,
		Nip:          u.Lecturer.Nip,
		Name:         u.Lecturer.Name,
		Email:        u.Lecturer.Email,
		TotalStudent: u.Lecturer.TotalStudent,
		StudyProgram: dto.StudyProgramResponse{
			ID:     u.Lecturer.StudyProgramID,
			Name:   u.Lecturer.StudyProgram.Name,
			Degree: u.Lecturer.StudyProgram.Degree,
			Faculty: dto.FacultyResponse{
				ID:   u.Lecturer.StudyProgram.FacultyID,
				Name: u.Lecturer.StudyProgram.Faculty.Name,
			},
		},
	}
}

// global variable for track event websocket for online users in session
var sessionEvent *dto.SessionEventPublish

func (ss *sessionService) Start(ctx context.Context, thesisID string) (*dto.SessionResponse, error) {
	// get information user login
	token := ctx.Value("Authorization").(string)
	userIDString, err := ss.jwt.GetUserIDByToken(token)
	if err != nil {
		ss.logger.Error("failed to extract user_id from token",
			zap.String("access_token", token),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserIDFromToken
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		ss.logger.Error("failed to parse user_id to uuid",
			zap.String("user_id_raw", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrParseStringToUUID
	}
	user, found, err := ss.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		ss.logger.Error("failed to fetch user by id",
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserByID
	}
	if !found {
		ss.logger.Warn("user not found",
			zap.String("user_id", userIDString),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}

	// handle existing session
	existing, found, _ := ss.sessionRepo.GetActiveSessionByThesisID(ctx, nil, thesisID)
	if existing != nil && found {
		ss.logger.Info("active session already exists",
			zap.String("thesis_id", thesisID),
			zap.String("session_id", existing.ID.String()),
		)

		return nil, dto.ErrSessionAlreadyStarted
	}

	// get thesis for prepare start session
	thesis, found, err := ss.sessionRepo.GetThesisByID(ctx, nil, thesisID)
	if err != nil {
		ss.logger.Error("failed to fetch thesis by id",
			zap.String("thesis_id", thesisID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetThesisByID
	}
	if !found {
		ss.logger.Warn("thesis not found",
			zap.String("thesis_id", thesisID),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}
	tID, err := uuid.Parse(thesisID)
	if err != nil {
		ss.logger.Error("failed to parse thesis_id to uuid",
			zap.String("thesis_id_raw", thesisID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrParseStringToUUID
	}

	// create session object
	sessionID := uuid.New()
	session := &entity.Session{
		ID:          sessionID,
		Status:      "waiting",
		UserIDOwner: userID,
		ThesisID:    tID,
	}
	// create session instance
	err = ss.sessionRepo.CreateSession(ctx, nil, session)
	if err != nil {
		ss.logger.Error("failed to create session",
			zap.String("session_id", sessionID.String()),
			zap.String("thesis_id", thesisID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrCreateSession
	}

	// determination of receiver and starter
	var (
		starter     string
		receiverIDs []uuid.UUID
	)
	switch {
	case user.StudentID != nil && *user.StudentID == thesis.StudentID:
		starter = user.Student.Name
		for _, sup := range thesis.Supervisors {
			receiverIDs = append(receiverIDs, sup.LecturerID)
		}
	case user.LecturerID != nil:
		starter = user.Lecturer.Name
		if thesis.StudentID != uuid.Nil {
			receiverIDs = append(receiverIDs, thesis.StudentID)
		}
		for _, sup := range thesis.Supervisors {
			if user.LecturerID != nil && sup.LecturerID != *user.LecturerID {
				receiverIDs = append(receiverIDs, sup.LecturerID)
			}
		}
	default:
		ss.logger.Warn("unauthorized attempt to start session",
			zap.String("user_id", userIDString),
			zap.String("thesis_id", thesisID),
		)
		return &dto.SessionResponse{}, errors.New("unauthorized: user not related to thesis")
	}

	// resolve receiver entity IDs (student/lecturer) -> user.id
	for _, rid := range receiverIDs {
		receiverUser, found, err := ss.userRepo.GetUserByStudentOrLecturerID(ctx, nil, rid.String())
		if err != nil {
			ss.logger.Error("failed to resolve receiver user",
				zap.String("receiver_entity_id", rid.String()),
				zap.Error(err),
			)
			continue // skip user ini tapi lanjut ke receiver lain
		}
		if !found {
			ss.logger.Warn("receiver user not found for entity_id",
				zap.String("receiver_entity_id", rid.String()),
			)
			continue
		}

		// create session event
		sessionEvent = &dto.SessionEventPublish{
			Event:    "session_started",
			ThesisID: tID,
		}

		if user.Role == constants.ENUM_ROLE_STUDENT {
			sessionEvent.StudentID = &thesis.StudentID
			sessionEvent.StudentName = thesis.Student.Name
		} else {
			for _, sup := range thesis.Supervisors {
				if user.LecturerID != nil && sup.LecturerID == *user.LecturerID {
					sessionEvent.Supervisors = append(sessionEvent.Supervisors, &dto.SessionSupervisor{
						ID:   sup.ID,
						Role: sup.Role,
						Name: user.Lecturer.Name,
					})
				}
			}
		}

		data, _ := json.Marshal(sessionEvent)

		// Send via WebSocket if online
		err = ss.wsService.SendToUser(receiverUser.ID.String(), data)
		if err != nil {
			// if failed / user offline → create new instance to notification table
			notif := &entity.Notification{
				ID:      uuid.New(),
				Title:   "New Thesis Session",
				Message: fmt.Sprintf("Your thesis session has been started by %s.", starter),
				IsRead:  false,
				UserID:  receiverUser.ID,
			}
			if err := ss.notificationRepo.CreateNotification(ctx, nil, notif); err != nil {
				ss.logger.Error("failed to create notification for offline user",
					zap.String("session_id", sessionID.String()),
					zap.String("thesis_id", thesisID),
					zap.String("receiver_id", rid.String()),
					zap.Error(err),
				)
				continue
			}
			ss.logger.Info("notification created for offline user",
				zap.String("session_id", sessionID.String()),
				zap.String("thesis_id", thesisID),
				zap.String("receiver_id", rid.String()),
			)
		} else {
			ss.logger.Info("session_started event sent via WebSocket",
				zap.String("session_id", sessionID.String()),
				zap.String("thesis_id", thesisID),
				zap.String("starter", starter),
				zap.String("receiver_id", rid.String()),
			)
		}
	}

	// create response
	res := &dto.SessionResponse{
		ID:     session.ID,
		Status: session.Status,
		Thesis: dto.ThesisResponse{
			ID:          thesis.ID,
			Title:       thesis.Title,
			Description: thesis.Description,
			Progress:    thesis.Progress,
			Student: &dto.CustomUserResponse{
				ID:         thesis.Student.ID,
				Name:       thesis.Student.Name,
				Identifier: thesis.Student.Nim,
			},
		},
		UserOwner: dto.UserResponse{
			ID:         user.ID,
			Identifier: user.Identifier,
			Role:       user.Role,
			Student:    mapStudent(*user),
			Lecturer:   mapLecturer(*user),
		},
	}
	for _, sup := range thesis.Supervisors {
		res.Thesis.Supervisors = append(res.Thesis.Supervisors, &dto.CustomUserResponse{
			ID:         sup.LecturerID,
			Name:       sup.Lecturer.Name,
			Identifier: sup.Lecturer.Nip,
		})
	}
	ss.logger.Info("new session started successfully",
		zap.String("session_id", sessionID.String()),
		zap.String("thesis_id", thesisID),
		zap.String("starter", starter),
	)

	return res, nil
}

func (ss *sessionService) Join(ctx context.Context, sessionID string) (*dto.SessionResponse, error) {
	// get information user login
	token := ctx.Value("Authorization").(string)
	userIDString, err := ss.jwt.GetUserIDByToken(token)
	if err != nil {
		ss.logger.Error("failed to extract user_id from token",
			zap.String("access_token", token),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserIDFromToken
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		ss.logger.Error("failed to parse user_id to uuid",
			zap.String("user_id_raw", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrParseStringToUUID
	}
	user, found, err := ss.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		ss.logger.Error("failed to fetch user by id",
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserByID
	}
	if !found {
		ss.logger.Warn("user not found",
			zap.String("user_id", userIDString),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}

	// get session
	session, found, err := ss.sessionRepo.GetActiveSessionBySessionID(ctx, nil, sessionID)
	if err != nil {
		ss.logger.Error("failed to fetch active session by id",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetActiveSessionBySessionID
	}
	if !found {
		ss.logger.Warn("active session not found",
			zap.String("session_id", sessionID),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}

	// cannot join if session is finished
	if session.Status == "finished" {
		ss.logger.Error("failed to join session because session is finished",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrSessionFinished
	}

	// same user id cannot start and join session
	if session.UserIDOwner == userID {
		ss.logger.Error("failed unable start and join session with the same user",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrUnableStartAndJoinSessionWithTheSameUser
	}

	now := time.Now()
	if session.Status == "waiting" {
		session.StartTime = &now
	}

	var receiverIDs []uuid.UUID
	update := false
	eventName := ""
	joiner := ""
	log.Println(user.Role)
	if user.StudentID != nil && session.Thesis.StudentID != uuid.Nil && *user.StudentID == session.Thesis.StudentID {
		update = true
		eventName = constants.ENUM_ROLE_STUDENT
		joiner = user.Student.Name
		for _, sup := range session.Thesis.Supervisors {
			receiverIDs = append(receiverIDs, sup.LecturerID)
		}
	} else if user.Role == constants.ENUM_ROLE_LECTURER {
		joiner = user.Lecturer.Name
		if session.UserOwner.Role == constants.ENUM_ROLE_STUDENT {
			update = true
		}
		if session.Thesis.StudentID != uuid.Nil {
			receiverIDs = append(receiverIDs, session.Thesis.StudentID)
		}
		for _, sup := range session.Thesis.Supervisors {
			if sup.LecturerID != user.Lecturer.ID {
				receiverIDs = append(receiverIDs, sup.LecturerID)
			} else {
				eventName = string(sup.Role)
			}
		}
	} else {
		return nil, dto.ErrUnauthorized
	}

	// resolve receiver entity IDs (student/lecturer) -> user.id
	for _, rid := range receiverIDs {
		receiverUser, found, err := ss.userRepo.GetUserByStudentOrLecturerID(ctx, nil, rid.String())
		if err != nil {
			ss.logger.Error("failed to resolve receiver user",
				zap.String("receiver_entity_id", rid.String()),
				zap.Error(err),
			)
			continue // skip user ini tapi lanjut ke receiver lain
		}
		if !found {
			ss.logger.Warn("receiver user not found for entity_id",
				zap.String("receiver_entity_id", rid.String()),
			)
			continue
		}

		// create session event
		sessionEvent.Event = fmt.Sprintf("%s_joined", eventName)
		if user.Role == constants.ENUM_ROLE_STUDENT {
			sessionEvent.StudentID = &session.Thesis.StudentID
			sessionEvent.StudentName = session.Thesis.Student.Name
		} else {
			if user.LecturerID != nil {
				for _, sup := range session.Thesis.Supervisors {
					if sup.LecturerID == *user.LecturerID {
						alreadyExists := false
						for _, existing := range sessionEvent.Supervisors {
							if existing.ID == sup.ID {
								alreadyExists = true
								break
							}
						}
						if !alreadyExists {
							sessionEvent.Supervisors = append(sessionEvent.Supervisors, &dto.SessionSupervisor{
								ID:   sup.ID,
								Role: sup.Role,
								Name: user.Lecturer.Name,
							})
						}
					}
				}
			}
		}

		data, _ := json.Marshal(sessionEvent)

		// Send via WebSocket if online
		err = ss.wsService.SendToUser(receiverUser.ID.String(), data)
		if err != nil {
			// if failed / user offline → create new instance to notification table
			notif := &entity.Notification{
				ID:      uuid.New(),
				Title:   "User has been join the session",
				Message: fmt.Sprintf("%s has joined the session.", joiner),
				IsRead:  false,
				UserID:  receiverUser.ID,
			}
			if err := ss.notificationRepo.CreateNotification(ctx, nil, notif); err != nil {
				ss.logger.Error("failed to create notification for offline user",
					zap.String("session_id", sessionID),
					zap.String("thesis_id", session.ThesisID.String()),
					zap.String("starter_id", session.UserIDOwner.String()),
					zap.Error(err),
				)
				continue
			}
			ss.logger.Info("notification created for offline user",
				zap.String("session_id", sessionID),
				zap.String("thesis_id", session.ThesisID.String()),
				zap.String("starter_id", session.UserIDOwner.String()),
			)
		} else {
			ss.logger.Info("user_joined event sent via WebSocket",
				zap.String("session_id", sessionID),
				zap.String("thesis_id", session.ThesisID.String()),
				zap.String("starter_id", session.UserIDOwner.String()),
			)
		}
	}

	if update {
		session.Status = "ongoing"
		// update session -> join & start session -> status = ongoing
		if err := ss.sessionRepo.UpdateSession(ctx, nil, session); err != nil {
			ss.logger.Error("failed to update session to ongoing",
				zap.String("session_id", sessionID),
				zap.Error(err),
			)
			return &dto.SessionResponse{}, dto.ErrUpdateSession
		}
		ss.logger.Info("session joined successfully, status updated to ongoing",
			zap.String("session_id", sessionID),
			zap.String("thesis_id", session.ThesisID.String()),
			zap.Time("start_time", now),
		)
	}

	res := &dto.SessionResponse{
		ID:        session.ID,
		StartTime: session.StartTime,
		Status:    session.Status,
		Thesis: dto.ThesisResponse{
			ID:          session.ThesisID,
			Title:       session.Thesis.Title,
			Description: session.Thesis.Description,
			Progress:    session.Thesis.Progress,
			Student: &dto.CustomUserResponse{
				ID:         session.Thesis.Student.ID,
				Name:       session.Thesis.Student.Name,
				Identifier: session.Thesis.Student.Nim,
			},
		},
		UserOwner: dto.UserResponse{
			ID:         session.UserOwner.ID,
			Identifier: session.UserOwner.Identifier,
			Role:       session.UserOwner.Role,
			Student:    mapStudent(session.UserOwner),
			Lecturer:   mapLecturer(session.UserOwner),
		},
	}
	for _, sup := range session.Thesis.Supervisors {
		res.Thesis.Supervisors = append(res.Thesis.Supervisors, &dto.CustomUserResponse{
			ID:         sup.LecturerID,
			Name:       sup.Lecturer.Name,
			Identifier: sup.Lecturer.Nip,
		})
	}

	return res, nil
}

func (ss *sessionService) Leave(ctx context.Context, sessionID string) (*dto.SessionResponse, error) {
	// get information user login
	token := ctx.Value("Authorization").(string)
	userIDString, err := ss.jwt.GetUserIDByToken(token)
	if err != nil {
		ss.logger.Error("failed to extract user_id from token",
			zap.String("access_token", token),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserIDFromToken
	}
	user, found, err := ss.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		ss.logger.Error("failed to fetch user by id",
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserByID
	}
	if !found {
		ss.logger.Warn("user not found",
			zap.String("user_id", userIDString),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}

	// get session
	session, found, err := ss.sessionRepo.GetActiveSessionBySessionID(ctx, nil, sessionID)
	if err != nil {
		ss.logger.Error("failed to fetch active session by id",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetActiveSessionBySessionID
	}
	if !found {
		ss.logger.Warn("active session not found",
			zap.String("session_id", sessionID),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}

	// cannot leave if session is not ongoing
	if session.Status == "waiting" {
		ss.logger.Error("failed to leave session because session has not started yet",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrSessionWaiting
	}
	if session.Status == "finished" {
		ss.logger.Error("failed to leave session because session is finished",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrSessionFinished
	}

	// cannot leave if owner session
	if user.ID == session.UserIDOwner {
		ss.logger.Error("failed leave because owner session",
			zap.String("session_id", sessionID),
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrOwnerSessionLeave
	}

	var receiverIDs []uuid.UUID
	eventName := ""
	leaver := ""
	if user.StudentID != nil && session.Thesis.StudentID != uuid.Nil && *user.StudentID == session.Thesis.StudentID {
		eventName = constants.ENUM_ROLE_STUDENT
		leaver = user.Student.Name
		for _, sup := range session.Thesis.Supervisors {
			receiverIDs = append(receiverIDs, sup.LecturerID)
		}
	} else if user.Role == constants.ENUM_ROLE_LECTURER {
		leaver = user.Lecturer.Name
		if session.Thesis.StudentID != uuid.Nil {
			receiverIDs = append(receiverIDs, session.Thesis.StudentID)
		}
		for _, sup := range session.Thesis.Supervisors {
			if sup.LecturerID != user.Lecturer.ID {
				receiverIDs = append(receiverIDs, sup.LecturerID)
			} else {
				eventName = string(sup.Role)
			}
		}
	} else {
		return nil, dto.ErrUnauthorized
	}

	// resolve receiver entity IDs (student/lecturer) -> user.id
	for _, rid := range receiverIDs {
		receiverUser, found, err := ss.userRepo.GetUserByStudentOrLecturerID(ctx, nil, rid.String())
		if err != nil {
			ss.logger.Error("failed to resolve receiver user",
				zap.String("receiver_entity_id", rid.String()),
				zap.Error(err),
			)
			continue // skip user ini tapi lanjut ke receiver lain
		}
		if !found {
			ss.logger.Warn("receiver user not found for entity_id",
				zap.String("receiver_entity_id", rid.String()),
			)
			continue
		}

		// create session event
		sessionEvent.Event = fmt.Sprintf("%s_leaved", eventName)
		if user.Role == constants.ENUM_ROLE_STUDENT {
			sessionEvent.StudentID = nil
			sessionEvent.StudentName = ""
		} else {
			if user.LecturerID != nil {
				for _, sup := range session.Thesis.Supervisors {
					if sup.LecturerID == *user.LecturerID {
						for j, existing := range sessionEvent.Supervisors {
							if existing.ID == sup.ID {
								sessionEvent.Supervisors = append(sessionEvent.Supervisors[:j], sessionEvent.Supervisors[j+1:]...)
								break
							}
						}
						break
					}
				}
			}
		}

		data, _ := json.Marshal(sessionEvent)

		// Send via WebSocket if online
		err = ss.wsService.SendToUser(receiverUser.ID.String(), data)
		if err != nil {
			// if failed / user offline → create new instance to notification table
			notif := &entity.Notification{
				ID:      uuid.New(),
				Title:   "User has been leave the session",
				Message: fmt.Sprintf("%s has leaved the session.", leaver),
				IsRead:  false,
				UserID:  receiverUser.ID,
			}
			if err := ss.notificationRepo.CreateNotification(ctx, nil, notif); err != nil {
				ss.logger.Error("failed to create notification for offline user",
					zap.String("session_id", sessionID),
					zap.String("thesis_id", session.ThesisID.String()),
					zap.String("starter_id", session.UserIDOwner.String()),
					zap.Error(err),
				)
				continue
			}
			ss.logger.Info("notification created for offline user",
				zap.String("session_id", sessionID),
				zap.String("thesis_id", session.ThesisID.String()),
				zap.String("starter_id", session.UserIDOwner.String()),
			)
		} else {
			ss.logger.Info("user_leaved event sent via WebSocket",
				zap.String("session_id", sessionID),
				zap.String("thesis_id", session.ThesisID.String()),
				zap.String("starter_id", session.UserIDOwner.String()),
			)
		}
	}

	res := &dto.SessionResponse{
		ID:        session.ID,
		StartTime: session.StartTime,
		Status:    session.Status,
		Thesis: dto.ThesisResponse{
			ID:          session.ThesisID,
			Title:       session.Thesis.Title,
			Description: session.Thesis.Description,
			Progress:    session.Thesis.Progress,
			Student: &dto.CustomUserResponse{
				ID:         session.Thesis.Student.ID,
				Name:       session.Thesis.Student.Name,
				Identifier: session.Thesis.Student.Nim,
			},
		},
		UserOwner: dto.UserResponse{
			ID:         session.UserOwner.ID,
			Identifier: session.UserOwner.Identifier,
			Role:       session.UserOwner.Role,
			Student:    mapStudent(session.UserOwner),
			Lecturer:   mapLecturer(session.UserOwner),
		},
	}
	for _, sup := range session.Thesis.Supervisors {
		res.Thesis.Supervisors = append(res.Thesis.Supervisors, &dto.CustomUserResponse{
			ID:         sup.LecturerID,
			Name:       sup.Lecturer.Name,
			Identifier: sup.Lecturer.Nip,
		})
	}

	return res, nil
}

func (ss *sessionService) End(ctx context.Context, sessionID string) (*dto.SessionResponse, error) {
	// get information user login
	token := ctx.Value("Authorization").(string)
	userIDString, err := ss.jwt.GetUserIDByToken(token)
	if err != nil {
		ss.logger.Error("failed to extract user_id from token",
			zap.String("access_token", token),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserIDFromToken
	}
	user, found, err := ss.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		ss.logger.Error("failed to fetch user by id",
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetUserByID
	}
	if !found {
		ss.logger.Warn("user not found",
			zap.String("user_id", userIDString),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}

	// get session
	session, found, err := ss.sessionRepo.GetActiveSessionBySessionID(ctx, nil, sessionID)
	if err != nil {
		ss.logger.Error("failed to fetch active session by id",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrGetActiveSessionBySessionID
	}
	if !found {
		ss.logger.Warn("active session not found",
			zap.String("session_id", sessionID),
		)
		return &dto.SessionResponse{}, dto.ErrNotFound
	}

	// only can end if still ongoing
	if session.Status == "waiting" {
		ss.logger.Error("failed to end session because session has not started yet",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrSessionWaiting
	}
	if session.Status == "finished" {
		ss.logger.Error("failed to end session because session is finished",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrSessionFinished
	}

	// only user id start can end session
	if session.UserIDOwner != user.ID {
		ss.logger.Error("failed to end session because not session owner",
			zap.String("session_id", sessionID),
			zap.String("user_id_start", session.UserIDOwner.String()),
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrNotOwnerSession
	}

	now := time.Now()
	session.Status = "finished"
	session.EndTime = &now
	// update session -> join & start session -> status = ongoing
	if err := ss.sessionRepo.UpdateSession(ctx, nil, session); err != nil {
		ss.logger.Error("failed to update session to finished",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return &dto.SessionResponse{}, dto.ErrUpdateSession
	}
	ss.logger.Info("session ended successfully, status updated to finished",
		zap.String("session_id", sessionID),
		zap.String("thesis_id", session.ThesisID.String()),
		zap.Time("end_time", now),
	)

	var receiverIDs []uuid.UUID
	ender := ""
	if user.StudentID != nil && session.Thesis.StudentID != uuid.Nil && *user.StudentID == session.Thesis.StudentID {
		ender = user.Student.Name
		for _, sup := range session.Thesis.Supervisors {
			receiverIDs = append(receiverIDs, sup.LecturerID)
		}
	} else if user.Role == constants.ENUM_ROLE_LECTURER {
		ender = user.Lecturer.Name
		if session.Thesis.StudentID != uuid.Nil {
			receiverIDs = append(receiverIDs, session.Thesis.StudentID)
		}
		for _, sup := range session.Thesis.Supervisors {
			if sup.LecturerID != user.Lecturer.ID {
				receiverIDs = append(receiverIDs, sup.LecturerID)
			}
		}
	} else {
		return nil, dto.ErrUnauthorized
	}

	// resolve receiver entity IDs (student/lecturer) -> user.id
	for _, rid := range receiverIDs {
		receiverUser, found, err := ss.userRepo.GetUserByStudentOrLecturerID(ctx, nil, rid.String())
		if err != nil {
			ss.logger.Error("failed to resolve receiver user",
				zap.String("receiver_entity_id", rid.String()),
				zap.Error(err),
			)
			continue // skip user ini tapi lanjut ke receiver lain
		}
		if !found {
			ss.logger.Warn("receiver user not found for entity_id",
				zap.String("receiver_entity_id", rid.String()),
			)
			continue
		}

		// create session event
		sessionEvent = &dto.SessionEventPublish{
			Event:    "user_ended",
			ThesisID: session.ThesisID,
		}

		data, _ := json.Marshal(sessionEvent)

		// Send via WebSocket if online
		err = ss.wsService.SendToUser(receiverUser.ID.String(), data)
		if err != nil {
			// if failed / user offline → create new instance to notification table
			notif := &entity.Notification{
				ID:      uuid.New(),
				Title:   "Session Ended",
				Message: fmt.Sprintf("%s has ended the session.", ender),
				IsRead:  false,
				UserID:  receiverUser.ID,
			}
			if err := ss.notificationRepo.CreateNotification(ctx, nil, notif); err != nil {
				ss.logger.Error("failed to create notification for offline user",
					zap.String("session_id", sessionID),
					zap.String("thesis_id", session.ThesisID.String()),
					zap.String("starter_id", session.UserIDOwner.String()),
					zap.Error(err),
				)
				continue
			}
			ss.logger.Info("notification created for offline user",
				zap.String("session_id", sessionID),
				zap.String("thesis_id", session.ThesisID.String()),
				zap.String("starter_id", session.UserIDOwner.String()),
			)
		} else {
			ss.logger.Info("user_ended event sent via WebSocket",
				zap.String("session_id", sessionID),
				zap.String("thesis_id", session.ThesisID.String()),
				zap.String("starter_id", session.UserIDOwner.String()),
			)
		}
	}

	res := &dto.SessionResponse{
		ID:        session.ID,
		StartTime: session.StartTime,
		EndTime:   session.EndTime,
		Status:    session.Status,
		Thesis: dto.ThesisResponse{
			ID:          session.ThesisID,
			Title:       session.Thesis.Title,
			Description: session.Thesis.Description,
			Progress:    session.Thesis.Progress,
			Student: &dto.CustomUserResponse{
				ID:         session.Thesis.Student.ID,
				Name:       session.Thesis.Student.Name,
				Identifier: session.Thesis.Student.Nim,
			},
		},
		UserOwner: dto.UserResponse{
			ID:         session.UserOwner.ID,
			Identifier: session.UserOwner.Identifier,
			Role:       session.UserOwner.Role,
			Student:    mapStudent(session.UserOwner),
			Lecturer:   mapLecturer(session.UserOwner),
		},
	}
	for _, sup := range session.Thesis.Supervisors {
		res.Thesis.Supervisors = append(res.Thesis.Supervisors, &dto.CustomUserResponse{
			ID:         sup.LecturerID,
			Name:       sup.Lecturer.Name,
			Identifier: sup.Lecturer.Nip,
		})
	}

	return res, nil
}

func (ss *sessionService) GetAll(ctx context.Context, filter dto.SessionFilterQuery) ([]*dto.SessionResponse, error) {
	token := ctx.Value("Authorization").(string)
	userIDString, err := ss.jwt.GetUserIDByToken(token)
	if err != nil {
		return nil, dto.ErrGetUserIDFromToken
	}
	user, found, err := ss.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		ss.logger.Error("failed to fetch user by id",
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return nil, dto.ErrGetUserByID
	}
	if !found {
		ss.logger.Warn("user not found",
			zap.String("user_id", userIDString),
		)
		return nil, dto.ErrNotFound
	}

	datas, err := ss.sessionRepo.GetAllSessionsByUserID(ctx, nil, user, filter)
	if err != nil {
		ss.logger.Error("failed to get all sessions",
			zap.Error(err),
		)
		return nil, dto.ErrGetAllSessionsByUserID
	}

	sessions := make([]*dto.SessionResponse, 0, len(datas))
	for _, data := range datas {
		session := &dto.SessionResponse{
			ID:        data.ID,
			StartTime: data.StartTime,
			EndTime:   data.EndTime,
			Status:    data.Status,
			Thesis: dto.ThesisResponse{
				ID:          data.ThesisID,
				Title:       data.Thesis.Title,
				Description: data.Thesis.Description,
				Progress:    data.Thesis.Progress,
				Student: &dto.CustomUserResponse{
					ID:         data.Thesis.Student.ID,
					Name:       data.Thesis.Student.Name,
					Identifier: data.Thesis.Student.Nim,
				},
			},
			UserOwner: dto.UserResponse{
				ID:         data.UserOwner.ID,
				Identifier: data.UserOwner.Identifier,
				Role:       data.UserOwner.Role,
				Student:    mapStudent(data.UserOwner),
				Lecturer:   mapLecturer(data.UserOwner),
			},
		}

		for _, sup := range data.Thesis.Supervisors {
			session.Thesis.Supervisors = append(session.Thesis.Supervisors, &dto.CustomUserResponse{
				ID:         sup.LecturerID,
				Name:       sup.Lecturer.Name,
				Identifier: sup.Lecturer.Nip,
			})
		}

		sessions = append(sessions, session)
	}
	ss.logger.Info("success get all sessions",
		zap.Int("count", len(datas)),
	)

	return sessions, nil
}

func (ss *sessionService) GetAllWithPagination(ctx context.Context, req response.PaginationRequest, filter dto.SessionFilterQuery) (dto.SessionPaginationResponse, error) {
	token := ctx.Value("Authorization").(string)
	userIDString, err := ss.jwt.GetUserIDByToken(token)
	if err != nil {
		return dto.SessionPaginationResponse{}, dto.ErrGetUserIDFromToken
	}
	user, found, err := ss.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		ss.logger.Error("failed to fetch user by id",
			zap.String("user_id", userIDString),
			zap.Error(err),
		)
		return dto.SessionPaginationResponse{}, dto.ErrGetUserByID
	}
	if !found {
		ss.logger.Warn("user not found",
			zap.String("user_id", userIDString),
		)
		return dto.SessionPaginationResponse{}, dto.ErrNotFound
	}

	datas, err := ss.sessionRepo.GetAllSessionsByUserIDWithPagination(ctx, nil, user, req, filter)
	if err != nil {
		ss.logger.Error("failed to get all sessions with pagination",
			zap.Error(err),
		)
		return dto.SessionPaginationResponse{}, dto.ErrGetAllSessionsByUserIDWithPagination
	}

	sessions := make([]*dto.SessionResponse, 0, len(datas.Sessions))
	for _, data := range datas.Sessions {
		session := &dto.SessionResponse{
			ID:        data.ID,
			StartTime: data.StartTime,
			EndTime:   data.EndTime,
			Status:    data.Status,
			Thesis: dto.ThesisResponse{
				ID:          data.ThesisID,
				Title:       data.Thesis.Title,
				Description: data.Thesis.Description,
				Progress:    data.Thesis.Progress,
				Student: &dto.CustomUserResponse{
					ID:         data.Thesis.Student.ID,
					Name:       data.Thesis.Student.Name,
					Identifier: data.Thesis.Student.Nim,
				},
			},
			UserOwner: dto.UserResponse{
				ID:         data.UserOwner.ID,
				Identifier: data.UserOwner.Identifier,
				Role:       data.UserOwner.Role,
				Student:    mapStudent(data.UserOwner),
				Lecturer:   mapLecturer(data.UserOwner),
			},
		}

		for _, sup := range data.Thesis.Supervisors {
			session.Thesis.Supervisors = append(session.Thesis.Supervisors, &dto.CustomUserResponse{
				ID:         sup.LecturerID,
				Name:       sup.Lecturer.Name,
				Identifier: sup.Lecturer.Nip,
			})
		}

		sessions = append(sessions, session)
	}
	ss.logger.Info("success get all sessions",
		zap.Int("count", len(datas.Sessions)),
	)

	return dto.SessionPaginationResponse{
		Data: sessions,
		PaginationResponse: response.PaginationResponse{
			Page:    datas.Page,
			PerPage: datas.PerPage,
			MaxPage: datas.MaxPage,
			Count:   datas.Count,
		},
	}, nil
}

func (ss *sessionService) GetDetail(ctx context.Context, id *string) (*dto.SessionResponse, error) {
	data, found, err := ss.sessionRepo.GetActiveSessionBySessionID(ctx, nil, *id)
	if err != nil {
		ss.logger.Error("failed to get session by id",
			zap.String("id", *id),
			zap.Error(err),
		)
		return nil, dto.ErrGetActiveSessionBySessionID
	}
	if !found {
		ss.logger.Warn("session not found",
			zap.String("id", *id),
		)
		return nil, dto.ErrNotFound
	}

	session := &dto.SessionResponse{
		ID:        data.ID,
		StartTime: data.StartTime,
		EndTime:   data.EndTime,
		Status:    data.Status,
		Thesis: dto.ThesisResponse{
			ID:          data.ThesisID,
			Title:       data.Thesis.Title,
			Description: data.Thesis.Description,
			Progress:    data.Thesis.Progress,
			Student: &dto.CustomUserResponse{
				ID:         data.Thesis.Student.ID,
				Name:       data.Thesis.Student.Name,
				Identifier: data.Thesis.Student.Nim,
			},
		},
		UserOwner: dto.UserResponse{
			ID:         data.UserOwner.ID,
			Identifier: data.UserOwner.Identifier,
			Role:       data.UserOwner.Role,
			Student:    mapStudent(data.UserOwner),
			Lecturer:   mapLecturer(data.UserOwner),
		},
	}

	for _, sup := range data.Thesis.Supervisors {
		session.Thesis.Supervisors = append(session.Thesis.Supervisors, &dto.CustomUserResponse{
			ID:         sup.LecturerID,
			Name:       sup.Lecturer.Name,
			Identifier: sup.Lecturer.Nip,
		})
	}
	ss.logger.Info("success get detail session",
		zap.String("id", *id),
	)

	return session, nil
}
