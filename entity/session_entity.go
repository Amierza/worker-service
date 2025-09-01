package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	StartTime time.Time     `json:"start_time"`
	EndTime   *time.Time    `json:"end_time"`
	Status    SessionStatus `gorm:"default:waiting" json:"status"`

	Notes    []Note    `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE;" json:"notes"`
	Messages []Message `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE;" json:"messages"`

	ThesisID uuid.UUID `gorm:"type:uuid;index" json:"thesis_id,omitempty"`
	Thesis   Thesis    `gorm:"foreignKey:ThesisID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"thesis,omitempty"`

	TimeStamp
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	var err error

	if !IsValidSessionStatus(s.Status) {
		return err
	}

	return nil
}
