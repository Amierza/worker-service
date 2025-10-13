package repository

import (
	"context"
	"errors"

	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

type (
	IUserRepository interface {
		// CREATE / POST

		// READ / GET
		GetUserByID(ctx context.Context, tx *gorm.DB, id string) (*entity.User, bool, error)
		GetUserByStudentOrLecturerID(ctx context.Context, tx *gorm.DB, targetID string) (*entity.User, bool, error)

		// UPDATE / PATCH

		// DELETE / DELETE
	}

	userRepository struct {
		db *gorm.DB
	}
)

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

// CREATE / POST

// READ / GET
func (nr *userRepository) GetUserByID(ctx context.Context, tx *gorm.DB, id string) (*entity.User, bool, error) {
	if tx == nil {
		tx = nr.db
	}

	var user *entity.User
	err := tx.WithContext(ctx).
		Preload("Messages").
		Preload("Notifications").
		Preload("SessionOwners").
		Preload("Student.StudyProgram.Faculty").
		Preload("Student.Theses").
		Preload("Lecturer.StudyProgram.Faculty").
		Preload("Lecturer.Supervisors").
		Where("id = ?", id).
		Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.User{}, false, nil
	}
	if err != nil {
		return &entity.User{}, false, err
	}

	return user, true, nil
}
func (ur *userRepository) GetUserByStudentOrLecturerID(ctx context.Context, tx *gorm.DB, targetID string) (*entity.User, bool, error) {
	if tx == nil {
		tx = ur.db
	}

	var user entity.User
	err := tx.WithContext(ctx).
		Preload("Messages").
		Preload("Notifications").
		Preload("SessionOwners").
		Preload("Student.StudyProgram.Faculty").
		Preload("Student.Theses").
		Preload("Lecturer.StudyProgram.Faculty").
		Preload("Lecturer.Supervisors").
		Where("student_id = ? OR lecturer_id = ?", targetID, targetID).
		First(&user).Error
	if err != nil {
		return &entity.User{}, false, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.User{}, false, nil
	}
	return &user, true, nil
}

// UPDATE / PATCH

// DELETE / DELETE
