package entity

import (
	"github.com/Amierza/chat-service/helper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Identifier string    `gorm:"not null" json:"identifier"`
	Role       Role      `gorm:"not null" json:"role"`
	Password   string    `json:"password"`

	Messages []Message `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE;" json:"messages"`

	StudentID *uuid.UUID `gorm:"type:uuid;index" json:"student_id,omitempty"`
	Student   Student    `gorm:"foreignKey:StudentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"student,omitempty"`

	LecturerID *uuid.UUID `gorm:"type:uuid;index" json:"lecturer_id,omitempty"`
	Lecturer   Lecturer   `gorm:"foreignKey:LecturerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"lecturer,omitempty"`

	TimeStamp
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	var err error

	u.Password, err = helper.HashPassword(u.Password)
	if err != nil {
		return err
	}

	if !IsValidRole(u.Role) {
		return err
	}

	return nil
}
