package usecase

import (
	"errors"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type BookmarkService struct {
	bookmarkRepo repository.IBookmarkRepository
	threadRepo   repository.IThreadRepository
	memberRepo   repository.ICommunityMemberRepository
}

func NewBookmarkService(bookmarkRepo repository.IBookmarkRepository, threadRepo repository.IThreadRepository, memberRepo repository.ICommunityMemberRepository) service.IBookmarkService {
	return &BookmarkService{bookmarkRepo: bookmarkRepo, threadRepo: threadRepo, memberRepo: memberRepo}
}

func (s *BookmarkService) Bookmark(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
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
	return s.bookmarkRepo.Save(&model.Bookmark{UserID: userID, ThreadID: threadID})
}

func (s *BookmarkService) Unbookmark(userID, threadID uint) error {
	exists, err := s.bookmarkRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not bookmarked")
	}
	return s.bookmarkRepo.Delete(userID, threadID)
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
