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
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	followsvciface "github.com/sunibgram/domain/service/follow"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

type FollowService struct {
	followRepo   followrepoiface.IFollowRepository
	userRepo     userrepoiface.IUserRepository
	blockRepo    userblockrepoiface.IUserBlockRepository
	cacheRepo    cacherepoiface.ICacheRepository
	notification notificationsvciface.INotificationService
}

func NewFollowService(
	followRepo followrepoiface.IFollowRepository,
	userRepo userrepoiface.IUserRepository,
	blockRepo userblockrepoiface.IUserBlockRepository,
	cacheRepo cacherepoiface.ICacheRepository,
	notification notificationsvciface.INotificationService,
) followsvciface.IFollowService {
	return &FollowService{followRepo: followRepo, userRepo: userRepo, blockRepo: blockRepo, cacheRepo: cacheRepo, notification: notification}
}

func (s *FollowService) Follow(followerID, followingID uint) error {
	if _, err := s.userRepo.FindByID(followingID); err != nil {
		return errors.New("user not found")
	}
	// Cannot follow someone in a mutual block relationship (re-follow guard).
	if s.blockRepo != nil {
		blocked, err := s.blockRepo.IsBlocked(followerID, followingID)
		if err != nil {
			return err
		}
		if blocked {
			return errors.New("blocked")
		}
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

func (s *FollowService) GetFollowing(viewerID, userID uint, page, limit int) ([]usermodel.User, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return nil, 0, errors.New("user not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	users, total, err := s.followRepo.GetFollowing(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return s.filterBlocked(viewerID, users), total, nil
}

func (s *FollowService) GetFollowers(viewerID, userID uint, page, limit int) ([]usermodel.User, int64, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return nil, 0, errors.New("user not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	users, total, err := s.followRepo.GetFollowers(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return s.filterBlocked(viewerID, users), total, nil
}

// filterBlocked drops users in a mutual block relationship with the viewer.
// Filtered at the service layer (not the repo query) because the follow-list
// cache is keyed on the profile owner, not the viewer — a viewer-specific WHERE
// would poison the shared cache. total stays approximate, same precedent as
// private-account filtering. Page-bounded (≤20) so the per-user IsBlocked loop
// is cheap.
func (s *FollowService) filterBlocked(viewerID uint, users []usermodel.User) []usermodel.User {
	if s.blockRepo == nil || viewerID == 0 || len(users) == 0 {
		return users
	}
	out := make([]usermodel.User, 0, len(users))
	for _, u := range users {
		if u.UserID == viewerID {
			out = append(out, u)
			continue
		}
		blocked, err := s.blockRepo.IsBlocked(viewerID, u.UserID)
		if err == nil && blocked {
			continue
		}
		out = append(out, u)
	}
	return out
}
