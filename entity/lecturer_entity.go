package entity

import (
	"github.com/google/uuid"
)

type Lecturer struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nip   string    `gorm:"unique;not null" json:"nip"`
	Name  string    `gorm:"not null" json:"name"`
	Email string    `gorm:"unique;not null" json:"email"`

	LecturerStudyPrograms []LecturerStudyProgram `gorm:"foreignKey:LecturerID;constraint:OnDelete:CASCADE;" json:"lecturer_study_programs"`

	FacultyID uuid.UUID `gorm:"type:uuid;index" json:"faculty_id,omitempty"`
	Faculty   Faculty   `gorm:"foreignKey:FacultyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"faculty,omitempty"`

	TimeStamp
}
