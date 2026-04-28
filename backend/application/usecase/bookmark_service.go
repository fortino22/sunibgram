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
}

func NewBookmarkService(bookmarkRepo repository.IBookmarkRepository, threadRepo repository.IThreadRepository) service.IBookmarkService {
	return &BookmarkService{bookmarkRepo: bookmarkRepo, threadRepo: threadRepo}
}

func (s *BookmarkService) Bookmark(userID, threadID uint) error {
	if _, err := s.threadRepo.FindByID(threadID); err != nil {
		return errors.New("thread not found")
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

func (s *BookmarkService) GetMyBookmarks(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.bookmarkRepo.FindByUserID(userID, page, limit)
}
