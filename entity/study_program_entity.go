package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudyProgram struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name   string    `gorm:"not null" json:"name"`
	Degree Degree    `gorm:"not null;default:s1" json:"degree"`

	Students  []Student  `gorm:"foreignKey:StudyProgramID;constraint:OnDelete:CASCADE;" json:"students"`
	Lecturers []Lecturer `gorm:"foreignKey:StudyProgramID;constraint:OnDelete:CASCADE;" json:"lecturers"`

	FacultyID uuid.UUID `gorm:"type:uuid;index" json:"faculty_id,omitempty"`
	Faculty   Faculty   `gorm:"foreignKey:FacultyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"faculty,omitempty"`

	TimeStamp
}

func (sp *StudyProgram) BeforeCreate(tx *gorm.DB) error {
	var err error

	if !IsValidDegree(sp.Degree) {
		return err
	}

	return nil
}
