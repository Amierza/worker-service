package entity

import (
	"github.com/Amierza/chat-service/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name     string    `gorm:"not null" json:"name"`
	Email    string    `gorm:"unique;not null" json:"email"`
	Password string    `json:"password"`

	Chatrooms []Chatroom `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"chatrooms"`

	TimeStamp
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	var err error

	u.Password, err = helper.HashPassword(u.Password)
	if err != nil {
		return err
	}

	return nil
}
