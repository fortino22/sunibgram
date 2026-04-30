package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ThreadService struct {
	threadRepo repository.IThreadRepository
}

func NewThreadService(threadRepo repository.IThreadRepository) service.IThreadService {
	return &ThreadService{threadRepo: threadRepo}
}

func (s *ThreadService) Create(ctx context.Context, userID uint, content string, parentID *uint) (*model.Thread, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is required")
	}

	var rootID *uint
	if parentID != nil {
		parent, err := s.threadRepo.FindByID(*parentID)
		if err != nil {
			return nil, errors.New("parent thread not found")
		}
		if parent.RootID == nil {
			rootID = &parent.ThreadID
		} else {
			rootID = parent.RootID
		}
	}

	thread := &model.Thread{
		UserID:   userID,
		Content:  content,
		ParentID: parentID,
		RootID:   rootID,
	}
	if err := s.threadRepo.Save(thread); err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *ThreadService) GetByID(id uint) (*model.Thread, error) {
	return s.threadRepo.FindByID(id)
}

func (s *ThreadService) GetRoots(page, limit int) ([]model.Thread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.threadRepo.FindRoots(page, limit)
}

func (s *ThreadService) GetChain(rootID uint, page, limit int) ([]model.Thread, int64, error) {
	root, err := s.threadRepo.FindByID(rootID)
	if err != nil {
		return nil, 0, errors.New("thread not found")
	}
	if root.ParentID != nil {
		return nil, 0, errors.New("thread is not a root post")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.threadRepo.FindByRootID(rootID, page, limit)
}

func (s *ThreadService) Update(ctx context.Context, id, userID uint, content string) (*model.Thread, error) {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("thread not found")
	}
	if thread.UserID != userID {
		return nil, errors.New("forbidden")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is required")
	}
	thread.Content = content
	if err := s.threadRepo.Update(thread); err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *ThreadService) GetMyPosts(userID uint, page, limit int) ([]model.Thread, []*time.Time, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.threadRepo.FindPostsByUserID(userID, page, limit)
}

func (s *ThreadService) GetMyReplies(userID uint, page, limit int) ([]model.Thread, []*time.Time, []*model.Thread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.threadRepo.FindReplyTabByUserID(userID, page, limit)
}

func (s *ThreadService) Delete(ctx context.Context, id, userID uint) error {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}
	return s.threadRepo.Delete(id)
}
