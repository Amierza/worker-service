package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entity.Student{},

		&entity.Chatroom{},
		&entity.Message{},

		&entity.Faculty{},
		&entity.StudyProgram{},
		&entity.Lecturer{},
		&entity.LecturerStudyProgram{},
	); err != nil {
		return err
	}

	return nil
}
