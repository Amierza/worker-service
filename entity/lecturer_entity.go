package entity

import (
	"github.com/google/uuid"
)

type Lecturer struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nip   string    `gorm:"unique;not null" json:"nip"`
	Name  string    `gorm:"not null" json:"name"`
	Email string    `gorm:"unique;not null" json:"email"`

	Users  []User   `gorm:"foreignKey:LecturerID;constraint:OnDelete:CASCADE;" json:"users"`
	Theses []Thesis `gorm:"foreignKey:LecturerID;constraint:OnDelete:CASCADE;" json:"thesises"`

	StudyProgramID uuid.UUID    `gorm:"type:uuid;index" json:"study_program_id,omitempty"`
	StudyProgram   StudyProgram `gorm:"foreignKey:StudyProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"study_program,omitempty"`

	TimeStamp
}
