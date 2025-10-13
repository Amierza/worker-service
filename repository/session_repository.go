package repository

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/entity"
	"github.com/Amierza/chat-service/response"
	"gorm.io/gorm"
)

type (
	ISessionRepository interface {
		// CREATE / POST
		CreateSession(ctx context.Context, tx *gorm.DB, session *entity.Session) error

		// READ / GET
		GetThesisByID(ctx context.Context, tx *gorm.DB, thesisID string) (*entity.Thesis, bool, error)
		GetActiveSessionByThesisID(ctx context.Context, tx *gorm.DB, thesisID string) (*entity.Session, bool, error)
		GetActiveSessionBySessionID(ctx context.Context, tx *gorm.DB, sessionID string) (*entity.Session, bool, error)
		GetAllSessionsByUserID(ctx context.Context, tx *gorm.DB, user *entity.User, filter dto.SessionFilterQuery) ([]*entity.Session, error)
		GetAllSessionsByUserIDWithPagination(ctx context.Context, tx *gorm.DB, user *entity.User, pagination response.PaginationRequest, filter dto.SessionFilterQuery) (dto.SessionPaginationRepositoryResponse, error)

		// UPDATE / PATCH
		UpdateSession(ctx context.Context, tx *gorm.DB, session *entity.Session) error

		// DELETE / DELETE
	}

	sessionRepository struct {
		db *gorm.DB
	}
)

func NewSessionRepository(db *gorm.DB) *sessionRepository {
	return &sessionRepository{
		db: db,
	}
}

// CREATE / POST
func (sr *sessionRepository) CreateSession(ctx context.Context, tx *gorm.DB, session *entity.Session) error {
	if tx == nil {
		tx = sr.db
	}

	return tx.WithContext(ctx).Create(&session).Error
}

// READ / GET
func (sr *sessionRepository) GetThesisByID(ctx context.Context, tx *gorm.DB, thesisID string) (*entity.Thesis, bool, error) {
	if tx == nil {
		tx = sr.db
	}

	thesis := &entity.Thesis{}
	err := tx.WithContext(ctx).
		Preload("ThesisLogs").
		Preload("Sessions").
		Preload("Supervisors.Lecturer.StudyProgram.Faculty").
		Preload("Student.StudyProgram.Faculty").
		Where("id = ?", thesisID).
		Take(&thesis).Error
	if err != nil {
		return &entity.Thesis{}, false, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.Thesis{}, false, nil
	}

	return thesis, true, nil
}
func (sr *sessionRepository) GetActiveSessionByThesisID(ctx context.Context, tx *gorm.DB, thesisID string) (*entity.Session, bool, error) {
	if tx == nil {
		tx = sr.db
	}

	session := &entity.Session{}
	err := tx.WithContext(ctx).
		Preload("Notes").
		Preload("Messages").
		Preload("Thesis.Supervisors.Lecturer.StudyProgram.Faculty").
		Preload("Thesis.Student.StudyProgram.Faculty").
		Preload("UserOwner.Student.StudyProgram.Faculty").
		Preload("UserOwner.Lecturer.StudyProgram.Faculty").
		Where("thesis_id = ? AND status != ?", thesisID, "finished").
		Take(&session).Error
	if err != nil {
		return &entity.Session{}, false, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.Session{}, false, nil
	}

	return session, true, nil
}
func (sr *sessionRepository) GetActiveSessionBySessionID(ctx context.Context, tx *gorm.DB, sessionID string) (*entity.Session, bool, error) {
	if tx == nil {
		tx = sr.db
	}

	session := &entity.Session{}
	err := tx.WithContext(ctx).
		Preload("Notes").
		Preload("Messages").
		Preload("Thesis.Supervisors.Lecturer.StudyProgram.Faculty").
		Preload("Thesis.Student.StudyProgram.Faculty").
		Preload("UserOwner.Student.StudyProgram.Faculty").
		Preload("UserOwner.Lecturer.StudyProgram.Faculty").
		Where("id = ?", sessionID).
		Take(&session).Error
	if err != nil {
		return &entity.Session{}, false, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.Session{}, false, nil
	}

	return session, true, nil
}
func (sr *sessionRepository) GetAllSessionsByUserID(ctx context.Context, tx *gorm.DB, user *entity.User, filter dto.SessionFilterQuery) ([]*entity.Session, error) {
	if tx == nil {
		tx = sr.db
	}

	var sessions []*entity.Session

	query := tx.WithContext(ctx).
		Model(&entity.Session{}).
		Joins("JOIN theses ON theses.id = sessions.thesis_id").
		Preload("Notes").
		Preload("Messages").
		Preload("Thesis.Supervisors.Lecturer.StudyProgram.Faculty").
		Preload("Thesis.Student.StudyProgram.Faculty").
		Preload("UserOwner.Student.StudyProgram.Faculty").
		Preload("UserOwner.Lecturer.StudyProgram.Faculty")

	// fFilter berdasarkan role (student / lecturer)
	if user.StudentID != nil {
		// mahasiswa: ambil semua session thesis miliknya
		query = query.Where("theses.student_id = ?", user.StudentID)
	} else if user.LecturerID != nil {
		// Dosen: ambil semua session thesis di mana dia menjadi supervisor
		subQuery := tx.
			Table("thesis_supervisors").
			Select("thesis_id").
			Where("lecturer_id = ?", user.LecturerID)

		query = query.Where("sessions.thesis_id IN (?)", subQuery)
	}

	// filter berdasarkan bulan (opsional)
	if filter.Month != "" {
		// Validasi: pastikan month angka 1-12
		monthInt, err := strconv.Atoi(filter.Month)
		if err == nil && monthInt >= 1 && monthInt <= 12 {
			query = query.Where("EXTRACT(MONTH FROM sessions.created_at) = ?", monthInt)
		}
	}

	// filter status
	switch filter.Status {
	case "waiting":
		query = query.Where("status = ?", "waiting")
	case "ongoing":
		query = query.Where("status = ?", "ongoing")
	case "finished":
		query = query.Where("status = ?", "finished")
	}

	// filter sorting
	switch filter.SortBy {
	case "latest":
		query = query.Order("created_at DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	}

	if err := query.Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}
func (sr *sessionRepository) GetAllSessionsByUserIDWithPagination(ctx context.Context, tx *gorm.DB, user *entity.User, pagination response.PaginationRequest, filter dto.SessionFilterQuery) (dto.SessionPaginationRepositoryResponse, error) {
	if tx == nil {
		tx = sr.db
	}

	var (
		sessions []*entity.Session
		err      error
		count    int64
	)

	if pagination.PerPage == 0 {
		pagination.PerPage = 10
	}

	if pagination.Page == 0 {
		pagination.Page = 1
	}

	query := tx.WithContext(ctx).
		Model(&entity.Session{}).
		Joins("JOIN theses ON theses.id = sessions.thesis_id").
		Preload("Notes").
		Preload("Messages").
		Preload("Thesis.Supervisors.Lecturer.StudyProgram.Faculty").
		Preload("Thesis.Student.StudyProgram.Faculty").
		Preload("UserOwner.Student.StudyProgram.Faculty").
		Preload("UserOwner.Lecturer.StudyProgram.Faculty")

	// cari berdasarkan role (student / lecturer)
	if user.StudentID != nil {
		// mahasiswa: ambil semua session thesis miliknya
		query = query.Where("theses.student_id = ?", user.StudentID)
	} else if user.LecturerID != nil {
		// Dosen: ambil semua session thesis di mana dia menjadi supervisor
		subQuery := tx.
			Table("thesis_supervisors").
			Select("thesis_id").
			Where("lecturer_id = ?", user.LecturerID)

		query = query.Where("sessions.thesis_id IN (?)", subQuery)
	}

	// filter berdasarkan bulan (opsional)
	if filter.Month != "" {
		// Validasi: pastikan month angka 1-12
		monthInt, err := strconv.Atoi(filter.Month)
		if err == nil && monthInt >= 1 && monthInt <= 12 {
			query = query.Where("EXTRACT(MONTH FROM sessions.created_at) = ?", monthInt)
		}
	}

	// filter status
	switch filter.Status {
	case "waiting":
		query = query.Where("status = waiting")
	case "ongoing":
		query = query.Where("status = ongoing")
	case "finished":
		query = query.Where("status = finished")
	}

	// filter sorting
	switch filter.SortBy {
	case "latest":
		query = query.Order("created_at DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	}

	if err := query.Count(&count).Error; err != nil {
		return dto.SessionPaginationRepositoryResponse{}, err
	}

	if err := query.Scopes(Paginate(pagination.Page, pagination.PerPage)).Find(&sessions).Error; err != nil {
		return dto.SessionPaginationRepositoryResponse{}, err
	}

	totalPage := int64(math.Ceil(float64(count) / float64(pagination.PerPage)))

	return dto.SessionPaginationRepositoryResponse{
		Sessions: sessions,
		PaginationResponse: response.PaginationResponse{
			Page:    pagination.Page,
			PerPage: pagination.PerPage,
			MaxPage: totalPage,
			Count:   count,
		},
	}, err
}

// UPDATE / PATCH
func (sr *sessionRepository) UpdateSession(ctx context.Context, tx *gorm.DB, session *entity.Session) error {
	if tx == nil {
		tx = sr.db
	}

	return tx.WithContext(ctx).Where("id = ?", session.ID).Updates(&session).Error
}
