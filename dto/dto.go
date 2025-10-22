package dto

import (
	"errors"
	"time"

	"github.com/Amierza/worker-service/entity"
	"github.com/google/uuid"
)

const (
	// ====================================== Failed ======================================
	// Token
	MESSAGE_FAILED_PROSES_REQUEST      = "failed proses request"
	MESSAGE_FAILED_ACCESS_DENIED       = "failed access denied"
	MESSAGE_FAILED_TOKEN_NOT_FOUND     = "failed token not found"
	MESSAGE_FAILED_TOKEN_NOT_VALID     = "failed token not valid"
	MESSAGE_FAILED_TOKEN_DENIED_ACCESS = "failed token denied access"

	// Consume
	FAILED_CONSUME_SUMMARY_TASKS = "failed consume summary tasks"

	// ====================================== Success ======================================
	// Consume
	SUCCESS_CONSUME_SUMMARY_TASKS = "success consume summary tasks"
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
	ThesisSummary struct {
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Progress    entity.Progress `json:"progress"`
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
		Role       string    `json:"role,omitempty"`
	}
)

// Task Summary Message
type (
	TaskSummary struct {
		SessionID     uuid.UUID  `json:"session_id"`
		SessionStatus string     `json:"session_status"`
		StartedAt     *time.Time `json:"started_at"`
		EndedAt       *time.Time `json:"ended_at"`
		CreatedAt     time.Time  `json:"created_at"`

		Owner       UserResponse       `json:"owner"`
		Student     StudentResponse    `json:"student"`
		Supervisors []LecturerResponse `json:"supervisors"`

		ThesisInfo ThesisSummary `json:"thesis_info"`

		Messages []MessageSummary `json:"messages"`
	}

	MessageSummary struct {
		ID              uuid.UUID          `json:"id"`
		IsText          bool               `json:"is_text"`
		Text            string             `json:"text,omitempty"`
		FileURL         string             `json:"file_url,omitempty"`
		FileType        string             `json:"file_type,omitempty"`
		Sender          CustomUserResponse `json:"sender"`
		ParentMessageID *uuid.UUID         `json:"parent_message_id,omitempty"`
		Timestamp       string             `json:"timestamp"`
	}
)
