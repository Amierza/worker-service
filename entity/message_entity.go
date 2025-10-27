package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	IsText  bool      `gorm:"not null" json:"is_text"`
	Text    string    `json:"text"`
	FileURL string    `json:"file_url,omitempty"`

	SenderRole Role      `gorm:"not null" json:"sender_role"`
	SenderID   uuid.UUID `gorm:"type:uuid;index" json:"sender_id"`
	Sender     User      `gorm:"foreignKey:SenderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sender"`

	SessionID uuid.UUID `gorm:"type:uuid;index" json:"session_id,omitempty"`
	Session   Session   `gorm:"foreignKey:SessionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"session,omitempty"`

	ParentMessageID *uuid.UUID `gorm:"type:uuid;index" json:"parent_message_id,omitempty"`

	TimeStamp
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	var err error

	if !IsValidRole(m.SenderRole) {
		return err
	}

	return nil
}
