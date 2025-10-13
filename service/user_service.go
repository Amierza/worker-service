package service

import (
	"context"

	"github.com/Amierza/chat-service/dto"
	"github.com/Amierza/chat-service/jwt"
	"github.com/Amierza/chat-service/repository"
	"go.uber.org/zap"
)

type (
	IUserService interface {
		GetProfile(ctx context.Context) (*dto.UserResponse, error)
	}

	userService struct {
		userRepo repository.IUserRepository
		logger   *zap.Logger
		jwt      jwt.IJWT
	}
)

func NewUserService(userRepo repository.IUserRepository, logger *zap.Logger, jwt jwt.IJWT) *userService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
		jwt:      jwt,
	}
}

func (us *userService) GetProfile(ctx context.Context) (*dto.UserResponse, error) {
	token := ctx.Value("Authorization").(string)
	userIDString, err := us.jwt.GetUserIDByToken(token)
	if err != nil {
		us.logger.Error("failed to get user_id by token",
			zap.String("id", userIDString),
			zap.Error(err),
		)
		return &dto.UserResponse{}, dto.ErrGetUserIDFromToken
	}

	data, found, err := us.userRepo.GetUserByID(ctx, nil, userIDString)
	if err != nil {
		us.logger.Error("failed to get user by id",
			zap.String("id", userIDString),
			zap.Error(err),
		)
		return nil, dto.ErrGetUserByID
	}
	if !found {
		us.logger.Warn("user not found",
			zap.String("id", userIDString),
		)
		return nil, dto.ErrNotFound
	}

	user := &dto.UserResponse{
		ID:         data.ID,
		Identifier: data.Identifier,
		Role:       data.Role,
	}
	if data.StudentID != nil {
		user.Student = &dto.StudentResponse{
			ID:    *data.StudentID,
			Nim:   data.Student.Nim,
			Name:  data.Student.Name,
			Email: data.Student.Email,
			StudyProgram: dto.StudyProgramResponse{
				ID:     data.Student.StudyProgramID,
				Name:   data.Student.StudyProgram.Name,
				Degree: data.Student.StudyProgram.Degree,
				Faculty: dto.FacultyResponse{
					ID:   data.Student.StudyProgram.Faculty.ID,
					Name: data.Student.StudyProgram.Faculty.Name,
				},
			},
		}

		// ambil 1 thesis (misal yang terakhir dibuat)
		if len(data.Student.Theses) > 0 {
			latestThesis := data.Student.Theses[len(data.Student.Theses)-1]
			user.ThesisID = latestThesis.ID.String()
		}
	}
	if data.LecturerID != nil {
		user.Lecturer = &dto.LecturerResponse{
			ID:           *data.LecturerID,
			Nip:          data.Lecturer.Nip,
			Name:         data.Lecturer.Name,
			Email:        data.Lecturer.Email,
			TotalStudent: data.Lecturer.TotalStudent,
			StudyProgram: dto.StudyProgramResponse{
				ID:     data.Lecturer.StudyProgramID,
				Name:   data.Lecturer.StudyProgram.Name,
				Degree: data.Lecturer.StudyProgram.Degree,
				Faculty: dto.FacultyResponse{
					ID:   data.Lecturer.StudyProgram.Faculty.ID,
					Name: data.Lecturer.StudyProgram.Faculty.Name,
				},
			},
		}
	}
	us.logger.Info("success get detail user",
		zap.String("id", userIDString),
	)

	return user, nil
}
