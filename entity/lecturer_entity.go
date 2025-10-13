package entity

import (
	"github.com/google/uuid"
)

type Lecturer struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nip          string    `gorm:"unique;not null" json:"nip"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"unique;not null" json:"email"`
	TotalStudent int       `json:"total_student"`

	Users       []User             `gorm:"foreignKey:LecturerID;constraint:OnDelete:CASCADE;" json:"users"`
	Supervisors []ThesisSupervisor `gorm:"foreignKey:LecturerID;constraint:OnDelete:CASCADE;" json:"supervisors"`

	StudyProgramID uuid.UUID    `gorm:"type:uuid;index" json:"study_program_id,omitempty"`
	StudyProgram   StudyProgram `gorm:"foreignKey:StudyProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"study_program,omitempty"`

	TimeStamp
}
