package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type BookmarkService struct {
	bookmarkRepo repository.IBookmarkRepository
	threadRepo   repository.IThreadRepository
	memberRepo   repository.ICommunityMemberRepository
	settingsRepo repository.IUserSettingsRepository
	followRepo   repository.IFollowRepository
	notification service.INotificationService
}

func NewBookmarkService(
	bookmarkRepo repository.IBookmarkRepository,
	threadRepo repository.IThreadRepository,
	memberRepo repository.ICommunityMemberRepository,
	settingsRepo repository.IUserSettingsRepository,
	followRepo repository.IFollowRepository,
	notification service.INotificationService,
) service.IBookmarkService {
	return &BookmarkService{
		bookmarkRepo: bookmarkRepo,
		threadRepo:   threadRepo,
		memberRepo:   memberRepo,
		settingsRepo: settingsRepo,
		followRepo:   followRepo,
		notification: notification,
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
		if _, err := s.memberRepo.FindMember(*thread.CommunityID, userID); err != nil {
			return errors.New("you must be a member of the community to bookmark this thread")
		}
	}
	exists, err := s.bookmarkRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already bookmarked")
	}
	if err := s.bookmarkRepo.Save(&model.Bookmark{UserID: userID, ThreadID: threadID}); err != nil {
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

func (s *BookmarkService) GetMyBookmarks(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.bookmarkRepo.FindByUserID(userID, page, limit)
}
