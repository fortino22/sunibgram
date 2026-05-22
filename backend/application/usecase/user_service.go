package usecase

import (
	"github.com/sunibgram/application/helper"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userprofilerepoiface "github.com/sunibgram/domain/repository/user_profile"
	usersvciface "github.com/sunibgram/domain/service/user"
)

type UserService struct {
	userRepo    userrepoiface.IUserRepository
	followRepo  followrepoiface.IFollowRepository
	profileRepo userprofilerepoiface.IUserProfileRepository
}

func NewUserService(userRepo userrepoiface.IUserRepository, followRepo followrepoiface.IFollowRepository, profileRepo userprofilerepoiface.IUserProfileRepository) usersvciface.IUserService {
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
