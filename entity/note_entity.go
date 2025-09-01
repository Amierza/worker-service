package entity

import (
	"github.com/google/uuid"
)

type Note struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Content string    `gorm:"not null" json:"content"`

	SessionID uuid.UUID `gorm:"type:uuid;index" json:"session_id,omitempty"`
	Session   Session   `gorm:"foreignKey:SessionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"session,omitempty"`

	TimeStamp
}
