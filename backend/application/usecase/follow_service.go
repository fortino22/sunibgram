package usecase

import (
	"errors"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type FollowService struct {
	followRepo repository.IFollowRepository
	userRepo   repository.IUserRepository
}

func NewFollowService(followRepo repository.IFollowRepository, userRepo repository.IUserRepository) service.IFollowService {
	return &FollowService{followRepo: followRepo, userRepo: userRepo}
}

func (s *FollowService) Follow(followerID, followingID uint) error {
	if _, err := s.userRepo.FindByID(followingID); err != nil {
		return errors.New("user not found")
	}
	return s.followRepo.Follow(followerID, followingID)
}

func (s *FollowService) Unfollow(followerID, followingID uint) error {
	return s.followRepo.Unfollow(followerID, followingID)
}

func (s *FollowService) GetStat(userID uint) (*model.UserStat, error) {
	return s.followRepo.GetStat(userID)
}
