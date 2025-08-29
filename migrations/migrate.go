package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entity.Faculty{},
		&entity.Lecturer{},
		&entity.StudyProgram{},
		&entity.Student{},
		&entity.LecturerStudyProgram{},
		&entity.User{},
		&entity.Chatroom{},
		&entity.Message{},
	); err != nil {
		return err
	}

	return nil
}
