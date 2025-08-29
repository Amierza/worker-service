package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Rollback(db *gorm.DB) error {
	tables := []interface{}{
		&entity.Message{},
		&entity.Chatroom{},
		&entity.User{},
		&entity.LecturerStudyProgram{},
		&entity.Student{},
		&entity.StudyProgram{},
		&entity.Lecturer{},
		&entity.Faculty{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}

	return nil
}
