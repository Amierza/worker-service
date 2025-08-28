package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	err := SeedFromJSON[entity.Student](db, "./migrations/json/students.json", entity.Student{}, "Email", "Nim")
	if err != nil {
		return err
	}

	return nil
}
