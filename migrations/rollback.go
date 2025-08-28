package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Rollback(db *gorm.DB) error {
	tables := []interface{}{
		&entity.LecturerStudyProgram{},
		&entity.Lecturer{},
		&entity.StudyProgram{},
		&entity.Faculty{},

		&entity.Message{},
		&entity.Chatroom{},

		&entity.Student{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}

	return nil
}
