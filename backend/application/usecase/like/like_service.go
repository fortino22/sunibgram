package likeuc

import (
	"context"
	"errors"
	"time"

	"github.com/sunibgram/application/helper"
	threaduc "github.com/sunibgram/application/usecase/thread"
	communitymodel "github.com/sunibgram/domain/model/community"
	likemodel "github.com/sunibgram/domain/model/like"
	threadmodel "github.com/sunibgram/domain/model/thread"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	likerepoiface "github.com/sunibgram/domain/repository/like"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	likesvciface "github.com/sunibgram/domain/service/like"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

type LikeService struct {
	likeRepo      likerepoiface.ILikeRepository
	threadRepo    threadrepoiface.IThreadRepository
	communityRepo communityrepoiface.ICommunityRepository
	memberRepo    communitymemberrepoiface.ICommunityMemberRepository
	settingsRepo  usersettingsrepoiface.IUserSettingsRepository
	followRepo    followrepoiface.IFollowRepository
	notification  notificationsvciface.INotificationService
}

func NewLikeService(
	likeRepo likerepoiface.ILikeRepository,
	threadRepo threadrepoiface.IThreadRepository,
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	settingsRepo usersettingsrepoiface.IUserSettingsRepository,
	followRepo followrepoiface.IFollowRepository,
	notification notificationsvciface.INotificationService,
) likesvciface.ILikeService {
	return &LikeService{
		likeRepo:      likeRepo,
		threadRepo:    threadRepo,
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		settingsRepo:  settingsRepo,
		followRepo:    followRepo,
		notification:  notification,
	}
}

func (s *LikeService) Like(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.Type == threadmodel.ThreadTypeRepost && thread.OriginalThreadID != nil {
		threadID = *thread.OriginalThreadID
		original, err := s.threadRepo.FindByID(threadID)
		if err != nil {
			return errors.New("thread not found")
		}
		thread = original
	}
	if err := threaduc.CanViewPrivateUser(context.Background(), s.settingsRepo, s.followRepo, &userID, thread.UserID); err != nil {
		return err
	}
	if thread.CommunityID != nil {
		community, err := s.communityRepo.FindByID(*thread.CommunityID)
		if err != nil {
			return errors.New("community not found")
		}
		if community.Visibility == communitymodel.CommunityVisibilityPrivate {
			if _, err := s.memberRepo.FindMember(*thread.CommunityID, userID); err != nil {
				return errors.New("you must be a member of the community to like this thread")
			}
		}
	}
	exists, err := s.likeRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already liked")
	}
	if err := s.likeRepo.Save(&likemodel.Like{UserID: userID, ThreadID: threadID}); err != nil {
		return err
	}
	s.threadRepo.InvalidateThread(threadID)
	helper.SendInteractionNotification(s.notification, "like", "liked your thread", userID, thread.UserID, threadID)
	return nil
}

func (s *LikeService) Unlike(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.Type == threadmodel.ThreadTypeRepost && thread.OriginalThreadID != nil {
		threadID = *thread.OriginalThreadID
		original, err := s.threadRepo.FindByID(threadID)
		if err != nil {
			return errors.New("thread not found")
		}
		thread = original
	}
	if err := threaduc.CanViewPrivateUser(context.Background(), s.settingsRepo, s.followRepo, &userID, thread.UserID); err != nil {
		return err
	}
	exists, err := s.likeRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not liked")
	}
	if err := s.likeRepo.Delete(userID, threadID); err != nil {
		return err
	}
	s.threadRepo.InvalidateThread(threadID)
	return nil
}

func (s *LikeService) IsLiked(userID, threadID uint) bool {
	ok, _ := s.likeRepo.Exists(userID, threadID)
	return ok
}

func (s *LikeService) AreThreadsLiked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return s.likeRepo.AreThreadsLiked(userID, threadIDs)
}

func (s *LikeService) AllLikedThreadIDs(userID uint) ([]uint, error) {
	return s.likeRepo.AllLikedThreadIDs(userID)
}

func (s *LikeService) GetMyLikes(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.likeRepo.FindByUserID(userID, page, limit)
}
