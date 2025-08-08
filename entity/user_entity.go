package entity

import (
	"github.com/Amierza/nawasena-backend/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Name     string    `json:"user_name"`
	Email    string    `gorm:"unique; not null" json:"user_email"`
	Password string    `json:"user_password"`

	TimeStamp
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	var err error
	u.Password, err = helper.HashPassword(u.Password)
	if err != nil {
		return err
	}

	return nil
}
