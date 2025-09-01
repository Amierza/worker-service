package entity

import (
	"github.com/google/uuid"
)

type Faculty struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `gorm:"unique;not null" json:"name"`

	StudyPrograms []StudyProgram `gorm:"foreignKey:FacultyID;constraint:OnDelete:CASCADE;" json:"study_programs"`

	TimeStamp
}
