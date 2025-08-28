package entity

import (
	"github.com/google/uuid"
)

type Message struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Sender  Sender    `gorm:"not null" json:"sender"`
	Content string    `json:"content"`

	ChatroomID uuid.UUID `gorm:"type:uuid;index" json:"chatroom_id,omitempty"`
	Chatroom   Chatroom  `gorm:"foreignKey:ChatroomID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"chatroom,omitempty"`

	TimeStamp
}
