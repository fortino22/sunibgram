package followuc

import (
	"context"
	"errors"

	"github.com/sunibgram/application/helper"
	communityuc "github.com/sunibgram/application/usecase/community"
	usermodel "github.com/sunibgram/domain/model/user"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	followsvciface "github.com/sunibgram/domain/service/follow"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

type FollowService struct {
	followRepo   followrepoiface.IFollowRepository
	userRepo     userrepoiface.IUserRepository
	cacheRepo    cacherepoiface.ICacheRepository
	notification notificationsvciface.INotificationService
}

func NewFollowService(
	followRepo followrepoiface.IFollowRepository,
	userRepo userrepoiface.IUserRepository,
	cacheRepo cacherepoiface.ICacheRepository,
	notification notificationsvciface.INotificationService,
) followsvciface.IFollowService {
	return &FollowService{followRepo: followRepo, userRepo: userRepo, cacheRepo: cacheRepo, notification: notification}
}

func (s *FollowService) Follow(followerID, followingID uint) error {
	if _, err := s.userRepo.FindByID(followingID); err != nil {
		return errors.New("user not found")
	}
	if err := s.followRepo.Follow(followerID, followingID); err != nil {
		return err
	}
	ctx := context.Background()
	communityuc.InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if s.cacheRepo != nil {
		_ = s.cacheRepo.InvalidateCommunityMembersGlobal(ctx)
	}
	helper.SendInteractionNotification(s.notification, "follow", "started following you", followerID, followingID, followingID)
	return nil
}

func (s *FollowService) Unfollow(followerID, followingID uint) error {
	if err := s.followRepo.Unfollow(followerID, followingID); err != nil {
		return err
	}
	ctx := context.Background()
	communityuc.InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if s.cacheRepo != nil {
		_ = s.cacheRepo.InvalidateCommunityMembersGlobal(ctx)
	}
	return nil
}

func (s *FollowService) GetStat(userID uint) (*userstatmodel.UserStat, error) {
	return s.followRepo.GetStat(userID)
}

func (s *FollowService) GetFollowing(userID uint, page, limit int) ([]usermodel.User, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return nil, 0, errors.New("user not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.followRepo.GetFollowing(userID, page, limit)
}

func (s *FollowService) GetFollowers(userID uint, page, limit int) ([]usermodel.User, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return nil, 0, errors.New("user not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.followRepo.GetFollowers(userID, page, limit)
}
