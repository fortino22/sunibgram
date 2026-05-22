package usecase

import (
	"github.com/sunibgram/application/helper"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
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

func (s *UserService) GetAllPaginated(page, limit int) ([]usermodel.User, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 10)
	return s.userRepo.FindAllPaginated(page, limit)
}

func (s *UserService) GetByID(id uint) (*usermodel.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) GetProfile(userID uint) (*usermodel.User, *userstatmodel.UserStat, *userprofilemodel.UserProfile, error) {
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
		profile = &userprofilemodel.UserProfile{UserID: userID}
	}
	return user, stat, profile, nil
}
