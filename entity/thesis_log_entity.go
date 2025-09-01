package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ThesisLog struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Progress Progress  `gorm:"default:bab1" json:"progress"`

	ThesisID uuid.UUID `gorm:"type:uuid;index" json:"thesis_id,omitempty"`
	Thesis   Thesis    `gorm:"foreignKey:ThesisID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"thesis,omitempty"`

	TimeStamp
}

func (tl *ThesisLog) BeforeCreate(tx *gorm.DB) error {
	var err error

	if !IsValidProgress(tl.Progress) {
		return err
	}

	return nil
}
