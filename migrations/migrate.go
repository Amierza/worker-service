package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entity.Faculty{},
		&entity.StudyProgram{},
		&entity.Student{},
		&entity.Lecturer{},
		&entity.User{},
		&entity.Notification{},
		&entity.Thesis{},
		&entity.ThesisSupervisor{},
		&entity.ThesisLog{},
		&entity.Session{},
		&entity.Message{},
		&entity.Note{},
	); err != nil {
		return err
	}

	return nil
}
