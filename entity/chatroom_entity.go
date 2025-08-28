package entity

import (
	"github.com/google/uuid"
)

type Chatroom struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title string    `gorm:"not null" json:"title"`

	Messages []Message `gorm:"foreignKey:ChatroomID;constraint:OnDelete:CASCADE;" json:"messages"`

	UserID uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`

	TimeStamp
}
