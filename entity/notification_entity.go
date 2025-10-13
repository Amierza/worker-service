package entity

import (
	"github.com/google/uuid"
)

type Notification struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title   string    `gorm:"not null" json:"title"`
	Message string    `gorm:"not null" json:"message"`
	IsRead  bool      `json:"is_read"`

	UserID uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`

	TimeStamp
}
