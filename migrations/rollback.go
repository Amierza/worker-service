package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Rollback(db *gorm.DB) error {
	tables := []interface{}{
		&entity.Note{},
		&entity.Message{},
		&entity.Session{},
		&entity.ThesisLog{},
		&entity.Thesis{},
		&entity.User{},
		&entity.Lecturer{},
		&entity.Student{},
		&entity.StudyProgram{},
		&entity.Faculty{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}

	return nil
}
