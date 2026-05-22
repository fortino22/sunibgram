package threadpinuc

import (
	"errors"

	threadpinmodel "github.com/sunibgram/domain/model/thread_pin"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	threadpinrepoiface "github.com/sunibgram/domain/repository/thread_pin"
	threadpinsvciface "github.com/sunibgram/domain/service/thread_pin"
)

type ThreadPinService struct {
	threadPinRepo threadpinrepoiface.IThreadPinRepository
	threadRepo    threadrepoiface.IThreadRepository
}

func NewThreadPinService(threadPinRepo threadpinrepoiface.IThreadPinRepository, threadRepo threadrepoiface.IThreadRepository) threadpinsvciface.IThreadPinService {
	return &ThreadPinService{threadPinRepo: threadPinRepo, threadRepo: threadRepo}
}

func (s *ThreadPinService) Pin(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}

	pinned, err := s.threadPinRepo.FindPinnedByUserID(userID)
	if err != nil {
		return err
	}
	if pinned != nil {
		if err := s.threadPinRepo.Unpin(userID, pinned.ThreadID); err != nil {
			return err
		}
	}

	return s.threadPinRepo.Pin(&threadpinmodel.ThreadPin{UserID: userID, ThreadID: threadID})
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

func (s *ThreadPinService) ArePinnedFor(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return s.threadPinRepo.ArePinnedFor(userID, threadIDs)
}

func (s *ThreadPinService) AllPinnedThreadIDs(userID uint) ([]uint, error) {
	return s.threadPinRepo.AllPinnedThreadIDs(userID)
}
