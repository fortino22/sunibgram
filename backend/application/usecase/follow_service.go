package usecase

import (
	"errors"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type FollowService struct {
	followRepo   repository.IFollowRepository
	userRepo     repository.IUserRepository
	notification service.INotificationService
}

func NewFollowService(
	followRepo repository.IFollowRepository,
	userRepo repository.IUserRepository,
	notification service.INotificationService,
) service.IFollowService {
	return &FollowService{followRepo: followRepo, userRepo: userRepo, notification: notification}
}

func (s *FollowService) Follow(followerID, followingID uint) error {
	if _, err := s.userRepo.FindByID(followingID); err != nil {
		return errors.New("user not found")
	}
	if err := s.followRepo.Follow(followerID, followingID); err != nil {
		return err
	}
	helper.SendInteractionNotification(s.notification, "follow", "started following you", followerID, followingID, followingID)
	return nil
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
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.followRepo.GetFollowing(userID, page, limit)
}

func (s *FollowService) GetFollowers(userID uint, page, limit int) ([]model.User, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return nil, 0, errors.New("user not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.followRepo.GetFollowers(userID, page, limit)
}
