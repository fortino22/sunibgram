package usecase

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type UserService struct {
	userRepo    repository.IUserRepository
	followRepo  repository.IFollowRepository
	profileRepo repository.IUserProfileRepository
}

func NewUserService(userRepo repository.IUserRepository, followRepo repository.IFollowRepository, profileRepo repository.IUserProfileRepository) service.IUserService {
	return &UserService{userRepo: userRepo, followRepo: followRepo, profileRepo: profileRepo}
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

func (s *UserService) GetProfile(userID uint) (*model.User, *model.UserStat, *model.UserProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, nil, err
	}
	stat, err := s.followRepo.GetStat(userID)
	if err != nil {
		return nil, nil, nil, err
	}
	profile, _ := s.profileRepo.FindByUserID(userID)
	if profile == nil {
		profile = &model.UserProfile{UserID: userID}
	}
	return user, stat, profile, nil
}
