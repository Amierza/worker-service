package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Thesis struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Progress    Progress  `json:"progress"`

	ThesisLogs  []ThesisLog        `gorm:"foreignKey:ThesisID;constraint:OnDelete:CASCADE;" json:"thesis_logs"`
	Sessions    []Session          `gorm:"foreignKey:ThesisID;constraint:OnDelete:CASCADE;" json:"sessions"`
	Supervisors []ThesisSupervisor `gorm:"foreignKey:ThesisID;constraint:OnDelete:CASCADE;" json:"supervisors"`

	StudentID uuid.UUID `gorm:"type:uuid;index" json:"student_id,omitempty"`
	Student   Student   `gorm:"foreignKey:StudentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"student,omitempty"`

	TimeStamp
}

func (t *Thesis) BeforeCreate(tx *gorm.DB) error {
	var err error

	if !IsValidProgress(t.Progress) {
		return err
	}

	return nil
}
