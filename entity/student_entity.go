package entity

import (
	"github.com/Amierza/chat-service/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Student struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nim      string    `gorm:"unique;not null" json:"nim"`
	Name     string    `gorm:"not null" json:"name"`
	Email    string    `gorm:"unique;not null" json:"email"`
	Password string    `json:"password"`

	Chatrooms []Chatroom `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE;" json:"chatrooms"`

	TimeStamp
}

func (s *Student) BeforeCreate(tx *gorm.DB) error {
	var err error

	s.Password, err = helper.HashPassword(s.Password)
	if err != nil {
		return err
	}

	return nil
}
