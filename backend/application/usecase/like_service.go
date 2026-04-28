package usecase

import (
	"errors"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type LikeService struct {
	likeRepo   repository.ILikeRepository
	threadRepo repository.IThreadRepository
}

func NewLikeService(likeRepo repository.ILikeRepository, threadRepo repository.IThreadRepository) service.ILikeService {
	return &LikeService{likeRepo: likeRepo, threadRepo: threadRepo}
}

func (s *LikeService) Like(userID, threadID uint) error {
	if _, err := s.threadRepo.FindByID(threadID); err != nil {
		return errors.New("thread not found")
	}
	exists, err := s.likeRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already liked")
	}
	return s.likeRepo.Save(&model.Like{UserID: userID, ThreadID: threadID})
}

func (s *LikeService) Unlike(userID, threadID uint) error {
	exists, err := s.likeRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not liked")
	}
	return s.likeRepo.Delete(userID, threadID)
}

func (s *LikeService) GetMyLikes(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.likeRepo.FindByUserID(userID, page, limit)
}
