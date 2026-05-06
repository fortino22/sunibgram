package usecase

import (
	"errors"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ThreadPinService struct {
	threadPinRepo repository.IThreadPinRepository
	threadRepo    repository.IThreadRepository
}

func NewThreadPinService(threadPinRepo repository.IThreadPinRepository, threadRepo repository.IThreadRepository) service.IThreadPinService {
	return &ThreadPinService{threadPinRepo: threadPinRepo, threadRepo: threadRepo}
}

func (s *ThreadPinService) Pin(userID, threadID uint) error {
	if _, err := s.threadRepo.FindByID(threadID); err != nil {
		return errors.New("thread not found")
	}
	exists, err := s.threadPinRepo.ExistsPin(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already pinned")
	}
	return s.threadPinRepo.Pin(&model.ThreadPin{UserID: userID, ThreadID: threadID})
}

func (s *ThreadPinService) Unpin(userID, threadID uint) error {
	exists, err := s.threadPinRepo.ExistsPin(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not pinned")
	}
	return s.threadPinRepo.Unpin(userID, threadID)
}

func (s *ThreadPinService) IsPinned(userID, threadID uint) bool {
	ok, _ := s.threadPinRepo.ExistsPin(userID, threadID)
	return ok
}

func (s *ThreadPinService) GetMyPins(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.threadPinRepo.FindByUserID(userID, page, limit)
}
