package entity

import (
	"github.com/google/uuid"
)

type LecturerStudyProgram struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	LecturerID     uuid.UUID    `gorm:"type:uuid;index" json:"lecturer_id,omitempty"`
	Lecturer       Lecturer     `gorm:"foreignKey:LecturerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"lecturer,omitempty"`
	StudyProgramID uuid.UUID    `gorm:"type:uuid;index" json:"study_program_id,omitempty"`
	StudyProgram   StudyProgram `gorm:"foreignKey:StudyProgramID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"study_program,omitempty"`

	TimeStamp
}
