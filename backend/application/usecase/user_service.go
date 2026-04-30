package usecase

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type UserService struct {
	userRepo   repository.IUserRepository
	followRepo repository.IFollowRepository
}

func NewUserService(userRepo repository.IUserRepository, followRepo repository.IFollowRepository) service.IUserService {
	return &UserService{userRepo: userRepo, followRepo: followRepo}
}

func (s *UserService) GetAllPaginated(page, limit int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.userRepo.FindAllPaginated(page, limit)
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) GetProfile(userID uint) (*model.User, *model.UserStat, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, err
	}
	stat, err := s.followRepo.GetStat(userID)
	if err != nil {
		return nil, nil, err
	}
	return user, stat, nil
}
