package entity

import (
	"github.com/google/uuid"
)

type Student struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nim   string    `gorm:"unique;not null" json:"nim"`
	Name  string    `gorm:"not null" json:"name"`
	Email string    `gorm:"unique;not null" json:"email"`

	FacultyID uuid.UUID `gorm:"type:uuid;index" json:"faculty_id,omitempty"`
	Faculty   Faculty   `gorm:"foreignKey:FacultyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"faculty,omitempty"`

	TimeStamp
}
