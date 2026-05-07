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

func (s *FollowService) GetFollowing(userID uint, page, limit int) ([]model.User, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return nil, 0, errors.New("user not found")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.followRepo.GetFollowing(userID, page, limit)
}

func (s *FollowService) GetFollowers(userID uint, page, limit int) ([]model.User, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return nil, 0, errors.New("user not found")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.followRepo.GetFollowers(userID, page, limit)
}
