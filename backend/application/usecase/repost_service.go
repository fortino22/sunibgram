package usecase

import (
	"errors"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type RepostService struct {
	repostRepo repository.IRepostRepository
	threadRepo repository.IThreadRepository
}

func NewRepostService(repostRepo repository.IRepostRepository, threadRepo repository.IThreadRepository) service.IRepostService {
	return &RepostService{repostRepo: repostRepo, threadRepo: threadRepo}
}

func (s *RepostService) Repost(userID, threadID uint) error {
	if _, err := s.threadRepo.FindByID(threadID); err != nil {
		return errors.New("thread not found")
	}
	exists, err := s.repostRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already reposted")
	}
	if err := s.repostRepo.Save(&model.Repost{UserID: userID, ThreadID: threadID}); err != nil {
		return err
	}
	return s.threadRepo.IncrementCount(threadID, "repost_count")
}

func (s *RepostService) Unrepost(userID, threadID uint) error {
	exists, err := s.repostRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not reposted")
	}
	if err := s.repostRepo.Delete(userID, threadID); err != nil {
		return err
	}
	return s.threadRepo.DecrementCount(threadID, "repost_count")
}

func (s *RepostService) GetMyReposts(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repostRepo.FindByUserID(userID, page, limit)
}
