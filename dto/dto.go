package dto

import (
	"errors"
	"time"

	"github.com/Amierza/chat-service/entity"
	"github.com/Amierza/chat-service/response"
	"github.com/google/uuid"
)

const (
	// ====================================== Failed ======================================
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"

	// Token
	MESSAGE_FAILED_PROSES_REQUEST      = "failed proses request"
	MESSAGE_FAILED_ACCESS_DENIED       = "failed access denied"
	MESSAGE_FAILED_TOKEN_NOT_FOUND     = "failed token not found"
	MESSAGE_FAILED_TOKEN_NOT_VALID     = "failed token not valid"
	MESSAGE_FAILED_TOKEN_DENIED_ACCESS = "failed token denied access"

	// Query Params
	MESSAGE_FAILED_INVALID_QUERY_PARAMS = "failed invalid query params"

	// Authentication
	FAILED_LOGIN         = "failed login"
	FAILED_REFRESH_TOKEN = "failed refresh token"

	// General Errors
	FAILED_CREATE      = "failed to create"
	FAILED_UPDATE      = "failed to update"
	FAILED_DELETE      = "failed to delete"
	FAILED_GET_ALL     = "failed to get all"
	FAILED_GET_DETAIL  = "failed to get detail"
	FAILED_GET_PROFILE = "failed to get profile"
	NOT_FOUND          = "not found"

	// Custom
	MESSAGE_FAILED_START_SESSION = "failed start session"
	MESSAGE_FAILED_JOIN_SESSION  = "failed join session"
	MESSAGE_FAILED_LEAVE_SESSION = "failed leave session"
	MESSAGE_FAILED_END_SESSION   = "failed end session"
	MESSAGE_FAILED_SEND_MESSAGE  = "failed send message"

	// ====================================== Success ======================================

	// Authentication
	SUCCESS_LOGIN         = "success login"
	SUCCESS_REFRESH_TOKEN = "success refresh token"

	// General Success
	SUCCESS_CREATE      = "success create"
	SUCCESS_UPDATE      = "success update"
	SUCCESS_DELETE      = "success delete"
	SUCCESS_GET_ALL     = "success get all"
	SUCCESS_GET_DETAIL  = "success get detail"
	SUCCESS_GET_PROFILE = "success to get profile"

	// Custom
	MESSAGE_SUCCESS_START_SESSION = "success start session"
	MESSAGE_SUCCESS_JOIN_SESSION  = "success join session"
	MESSAGE_SUCCESS_LEAVE_SESSION = "success leave session"
	MESSAGE_SUCCESS_END_SESSION   = "success end session"
	MESSAGE_SUCCESS_SEND_MESSAGE  = "success send message"
)

var (
	// Not Found
	ErrNotFound = errors.New("not found")

	// Unauthorized
	ErrUnauthorized = errors.New("unauthorized")

	// Token
	ErrGenerateAccessToken           = errors.New("failed to generate access token")
	ErrGenerateRefreshToken          = errors.New("failed to generate refresh token")
	ErrUnexpectedSigningMethod       = errors.New("unexpected signing method")
	ErrDecryptToken                  = errors.New("failed to decrypt token")
	ErrTokenInvalid                  = errors.New("token invalid")
	ErrValidateToken                 = errors.New("failed to validate token")
	ErrGetUserIDFromToken            = errors.New("failed get user id from token")
	ErrGetUserRoleFromToken          = errors.New("failed get user role from token")
	ErrGenerateAccessAndRefreshToken = errors.New("failed generate access and refresh token")

	// Redis
	ErrPushToRedis = errors.New("failed push to redis")

	// Parse
	ErrParseStringToUUID = errors.New("failed parse string to uuid format")
	ErrMarshalToJSON     = errors.New("failed marshal to JSON")

	// Authentication
	ErrInvalidEmail      = errors.New("email must be in a valid format (ex: user123@example.com)")
	ErrInvalidPassword   = errors.New("password must be at least 8 characters long")
	ErrIncorrectPassword = errors.New("incorrect password")

	// User
	ErrGetUserByIdentifier = errors.New("failed get user by identifier")
	ErrGetUserByID         = errors.New("failed get user by id")

	// Thesis
	ErrGetThesisByID = errors.New("failed get thesis by id")

	// Session
	ErrCreateSession                                = errors.New("failed create session")
	ErrUpdateSession                                = errors.New("failed update session")
	ErrGetActiveSessionBySessionID                  = errors.New("failed get active session by session id")
	ErrGetAllSessionsByUserID                       = errors.New("failed get all sessions by user id")
	ErrGetAllSessionsByUserIDWithPagination         = errors.New("failed get all sessions by user id with pagination")
	ErrSessionAlreadyStarted                        = errors.New("failed session already started")
	ErrUnableStartAndJoinSessionWithTheSameUser     = errors.New("failed start and join session with the same user")
	ErrOnlyStudentCanJoinLecturerWhenStartedSession = errors.New("failed only student can join lecturer when started session")
	ErrNotOwnerSession                              = errors.New("failed end session because not owner session")
	ErrStatusNotOngoing                             = errors.New("failed end session because not ongoing state")
	ErrOwnerSessionLeave                            = errors.New("failed leave because owner session")
	ErrSessionFinished                              = errors.New("failed session is finished")
	ErrInvalidSessionStatus                         = errors.New("failed invalid session status")
	ErrSessionWaiting                               = errors.New("failed session not started yet")

	// Notification
	ErrGetAllNotificationsByUserID = errors.New("failed get all notifications by user id")
	ErrGetNotificationByID         = errors.New("failed get notification by id")
	ErrUpdateIsReadNotification    = errors.New("failed update is_read notification")

	// Message
	ErrGetAllMessageWithPagination = errors.New("failed get all message with pagination")
)

// Master
type (
	FacultyResponse struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	StudyProgramResponse struct {
		ID      uuid.UUID       `json:"id"`
		Name    string          `json:"name"`
		Degree  entity.Degree   `json:"degree"`
		Faculty FacultyResponse `json:"faculty"`
	}
	StudentResponse struct {
		ID           uuid.UUID            `json:"id"`
		Nim          string               `json:"nim"`
		Name         string               `json:"name"`
		Email        string               `json:"email"`
		StudyProgram StudyProgramResponse `json:"study_program"`
	}
	LecturerResponse struct {
		ID           uuid.UUID            `json:"id"`
		Nip          string               `json:"nip"`
		Name         string               `json:"name"`
		Email        string               `json:"email"`
		TotalStudent int                  `json:"total_student"`
		StudyProgram StudyProgramResponse `json:"study_program"`
	}
	ThesisResponse struct {
		ID          uuid.UUID             `gorm:"type:uuid;primaryKey" json:"id"`
		Title       string                `gorm:"not null" json:"title"`
		Description string                `json:"description"`
		Progress    entity.Progress       `json:"progress"`
		Student     *CustomUserResponse   `json:"student,omitempty"`
		Supervisors []*CustomUserResponse `json:"supervisors,omitempty"`
	}
)

// Authentiation for Admin
type (
	LoginRequest struct {
		Identifier string `json:"identifier" binding:"required" example:"187xxxxxx"`
		Password   string `json:"password" binding:"required" example:"secret123"`
	}
	LoginResponse struct {
		AccessToken  string `json:"access_token" example:"<access_token_here>"`
		RefreshToken string `json:"refresh_token" example:"<refresh_token_here>"`
	}
	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required" example:"<refresh_token_here>"`
	}
	RefreshTokenResponse struct {
		AccessToken string `json:"access_token" example:"<new_access_token_here>"`
	}
)

// User
type (
	UserResponse struct {
		ID         uuid.UUID         `json:"id"`
		Identifier string            `json:"identifier"`
		Role       entity.Role       `json:"role"`
		ThesisID   string            `json:"thesis_id,omitempty"`
		Student    *StudentResponse  `json:"student,omitempty"`
		Lecturer   *LecturerResponse `json:"lecturer,omitempty"`
	}
	CustomUserResponse struct {
		ID         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Identifier string    `json:"identifier"`
	}
)

// Session
type (
	SessionResponse struct {
		ID        uuid.UUID            `json:"id"`
		StartTime *time.Time           `json:"start_time,omitempty"`
		EndTime   *time.Time           `json:"end_time,omitempty"`
		Status    entity.SessionStatus `json:"status"`
		Thesis    ThesisResponse       `json:"thesis"`
		UserOwner UserResponse         `json:"user_owner"`
	}
	// Filter
	SessionFilterQuery struct {
		SortBy string `form:"sort"` // ex: latest, oldest
		Status string `form:"status"`
		Month  string `form:"month"`
	}
	SessionSupervisor struct {
		ID   uuid.UUID   `json:"id"`
		Role entity.Role `json:"role"`
		Name string      `json:"name"`
	}
	SessionEventPublish struct {
		Event       string               `json:"event"`
		ThesisID    uuid.UUID            `json:"thesis_id"`
		StudentID   *uuid.UUID           `json:"student_id,omitempty"`
		StudentName string               `json:"student_name,omitempty"`
		Supervisors []*SessionSupervisor `json:"supervisors,omitempty"`
	}
	SessionPaginationResponse struct {
		response.PaginationResponse
		Data []*SessionResponse `json:"data"`
	}
	SessionPaginationRepositoryResponse struct {
		response.PaginationResponse
		Sessions []*entity.Session
	}
)

// Notification
type (
	NotificationResponse struct {
		ID      uuid.UUID `json:"id"`
		UserID  uuid.UUID `json:"user_id"`
		Title   string    `json:"title"`
		Message string    `json:"message"`
		IsRead  bool      `json:"is_read"`
	}
)

// Message
type (
	MessageResponse struct {
		ID              uuid.UUID   `json:"id"`
		IsText          bool        `json:"is_text"`
		Text            string      `json:"text"`
		FileURL         string      `json:"file_url,omitempty"`
		FileType        string      `json:"file_type,omitempty"`
		SenderRole      entity.Role `json:"sender_role"`
		SenderID        uuid.UUID   `json:"sender_id"`
		SenderName      string      `json:"sender_name"`
		SessionID       uuid.UUID   `json:"session_id"`
		ParentMessageID *uuid.UUID  `json:"parent_message_id,omitempty"`
	}
	MessageEventPublish struct {
		MessageID       uuid.UUID   `json:"id"`
		Event           string      `json:"event"`
		IsText          bool        `json:"is_text"`
		Text            string      `json:"text"`
		FileURL         string      `json:"file_url,omitempty"`
		FileType        string      `json:"file_type,omitempty"`
		SenderRole      entity.Role `json:"sender_role"`
		SenderID        uuid.UUID   `json:"sender_id"`
		SenderName      string      `json:"sender_name"`
		SessionID       uuid.UUID   `json:"session_id"`
		ParentMessageID *uuid.UUID  `json:"parent_message_id,omitempty"`
	}
	SendMessageRequest struct {
		IsText          bool       `json:"is_text" binding:"required"`
		Text            string     `json:"text" binding:"required"`
		FileURL         string     `json:"file_url,omitempty"`
		FileType        string     `json:"file_type,omitempty"`
		ParentMessageID *uuid.UUID `json:"parent_message_id,omitempty"`
	}
	MessagePaginationResponse struct {
		response.PaginationResponse
		Data []MessageResponse `json:"data"`
	}
	MessagePaginationRepositoryResponse struct {
		response.PaginationResponse
		Messages []entity.Message
	}
)
