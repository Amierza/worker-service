package entity

import (
	"github.com/google/uuid"
)

type Chatroom struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title string    `gorm:"not null" json:"title"`

	Messages []Message `gorm:"foreignKey:ChatroomID;constraint:OnDelete:CASCADE;" json:"messages"`

	StudentID uuid.UUID `gorm:"type:uuid;index" json:"student_id,omitempty"`
	Student   Student   `gorm:"foreignKey:StudentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"student,omitempty"`

	TimeStamp
}
