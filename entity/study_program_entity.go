package entity

import (
	"github.com/google/uuid"
)

type StudyProgram struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name   string    `gorm:"not null" json:"name"`
	Degree string    `gorm:"not null" json:"degree"`

	LecturerStudyPrograms []LecturerStudyProgram `gorm:"foreignKey:StudyProgramID;constraint:OnDelete:CASCADE;" json:"lecturer_study_programs"`

	FacultyID uuid.UUID `gorm:"type:uuid;index" json:"faculty_id,omitempty"`
	Faculty   Faculty   `gorm:"foreignKey:FacultyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"faculty,omitempty"`

	TimeStamp
}
