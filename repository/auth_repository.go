package repository

import (
	"context"
	"errors"

	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

type (
	IAuthRepository interface {
		GetUserByIdentifier(ctx context.Context, tx *gorm.DB, identifier *string) (*entity.User, bool, error)
	}

	authRepository struct {
		db *gorm.DB
	}
)

func NewAuthRepository(db *gorm.DB) *authRepository {
	return &authRepository{
		db: db,
	}
}

func (ar *authRepository) GetUserByIdentifier(ctx context.Context, tx *gorm.DB, identifier *string) (*entity.User, bool, error) {
	if tx == nil {
		tx = ar.db
	}

	var user *entity.User
	err := tx.WithContext(ctx).Where("identifier = ?", &identifier).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.User{}, false, nil
	}
	if err != nil {
		return &entity.User{}, false, err
	}

	return user, true, nil
}
