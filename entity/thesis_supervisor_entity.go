package entity

import (
	"github.com/google/uuid"
)

type ThesisSupervisor struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Role Role      `gorm:"not null" json:"role"`

	ThesisID uuid.UUID `gorm:"type:uuid;index" json:"thesis_id,omitempty"`
	Thesis   Thesis    `gorm:"foreignKey:ThesisID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"thesis,omitempty"`

	LecturerID uuid.UUID `gorm:"type:uuid;index" json:"lecturer_id,omitempty"`
	Lecturer   Lecturer  `gorm:"foreignKey:LecturerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"lecturer,omitempty"`

	TimeStamp
}
