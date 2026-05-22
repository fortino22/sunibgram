package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/sunibgram/application/helper"
	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
	bookmarkrepoiface "github.com/sunibgram/domain/repository/bookmark"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

type BookmarkService struct {
	bookmarkRepo  bookmarkrepoiface.IBookmarkRepository
	threadRepo    threadrepoiface.IThreadRepository
	communityRepo communityrepoiface.ICommunityRepository
	memberRepo    communitymemberrepoiface.ICommunityMemberRepository
	settingsRepo  usersettingsrepoiface.IUserSettingsRepository
	followRepo    followrepoiface.IFollowRepository
	notification  notificationsvciface.INotificationService
}

func NewBookmarkService(
	bookmarkRepo bookmarkrepoiface.IBookmarkRepository,
	threadRepo threadrepoiface.IThreadRepository,
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	settingsRepo usersettingsrepoiface.IUserSettingsRepository,
	followRepo followrepoiface.IFollowRepository,
	notification notificationsvciface.INotificationService,
) bookmarksvciface.IBookmarkService {
	return &BookmarkService{
		bookmarkRepo:  bookmarkRepo,
		threadRepo:    threadRepo,
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		settingsRepo:  settingsRepo,
		followRepo:    followRepo,
		notification:  notification,
	}
}

func (s *BookmarkService) Bookmark(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if err := canViewPrivateUser(context.Background(), s.settingsRepo, s.followRepo, &userID, thread.UserID); err != nil {
		return err
	}
	if thread.CommunityID != nil {
		community, err := s.communityRepo.FindByID(*thread.CommunityID)
		if err != nil {
			return errors.New("community not found")
		}
		if community.Visibility == communitymodel.CommunityVisibilityPrivate {
			if _, err := s.memberRepo.FindMember(*thread.CommunityID, userID); err != nil {
				return errors.New("you must be a member of the community to bookmark this thread")
			}
		}
	}
	exists, err := s.bookmarkRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already bookmarked")
	}
	if err := s.bookmarkRepo.Save(&bookmarkmodel.Bookmark{UserID: userID, ThreadID: threadID}); err != nil {
		return err
	}
	s.threadRepo.InvalidateThread(threadID)
	helper.SendInteractionNotification(s.notification, "bookmark", "bookmarked your thread", userID, thread.UserID, threadID)
	return nil
}

func (s *BookmarkService) Unbookmark(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if err := canViewPrivateUser(context.Background(), s.settingsRepo, s.followRepo, &userID, thread.UserID); err != nil {
		return err
	}
	exists, err := s.bookmarkRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not bookmarked")
	}
	if err := s.bookmarkRepo.Delete(userID, threadID); err != nil {
		return err
	}
	s.threadRepo.InvalidateThread(threadID)
	return nil
}

func (s *BookmarkService) IsBookmarked(userID, threadID uint) bool {
	ok, _ := s.bookmarkRepo.Exists(userID, threadID)
	return ok
}

func (s *BookmarkService) AreThreadsBookmarked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return s.bookmarkRepo.AreThreadsBookmarked(userID, threadIDs)
}

func (s *BookmarkService) AllBookmarkedThreadIDs(userID uint) ([]uint, error) {
	return s.bookmarkRepo.AllBookmarkedThreadIDs(userID)
}

func (s *BookmarkService) GetMyBookmarks(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.bookmarkRepo.FindByUserID(userID, page, limit)
}
