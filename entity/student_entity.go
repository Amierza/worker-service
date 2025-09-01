package entity

import (
	"github.com/google/uuid"
)

type Student struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nim   string    `gorm:"unique;not null" json:"nim"`
	Name  string    `gorm:"not null" json:"name"`
	Email string    `gorm:"unique;not null" json:"email"`

	Users  []User   `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE;" json:"users"`
	Theses []Thesis `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE;" json:"thesises"`

	StudyProgramID uuid.UUID    `gorm:"type:uuid;index" json:"study_program_id,omitempty"`
	StudyProgram   StudyProgram `gorm:"foreignKey:StudyProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"study_program,omitempty"`

	TimeStamp
}
