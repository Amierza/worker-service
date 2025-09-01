package migrations

import (
	"github.com/Amierza/chat-service/entity"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	err := SeedFromJSON[entity.Faculty](db, "./migrations/json/faculties.json", entity.Faculty{}, "Name")
	if err != nil {
		return err
	}

	err = SeedFromJSON[entity.StudyProgram](db, "./migrations/json/study_programs.json", entity.StudyProgram{}, "Name", "Degree")
	if err != nil {
		return err
	}

	err = SeedFromJSON[entity.Student](db, "./migrations/json/students.json", entity.Student{}, "Nim")
	if err != nil {
		return err
	}

	err = SeedFromJSON[entity.Lecturer](db, "./migrations/json/lecturers.json", entity.Lecturer{}, "Nip")
	if err != nil {
		return err
	}

	err = SeedFromJSON[entity.User](db, "./migrations/json/users.json", entity.User{}, "Identifier", "Role")
	if err != nil {
		return err
	}

	err = SeedFromJSON[entity.Thesis](db, "./migrations/json/theses.json", entity.Thesis{}, "StudentID", "LecturerID", "Title")
	if err != nil {
		return err
	}

	return nil
}
