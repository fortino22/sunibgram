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
	memberRepo repository.ICommunityMemberRepository
}

func NewLikeService(likeRepo repository.ILikeRepository, threadRepo repository.IThreadRepository, memberRepo repository.ICommunityMemberRepository) service.ILikeService {
	return &LikeService{likeRepo: likeRepo, threadRepo: threadRepo, memberRepo: memberRepo}
}

func (s *LikeService) Like(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.CommunityID != nil {
		if _, err := s.memberRepo.FindMember(*thread.CommunityID, userID); err != nil {
			return errors.New("you must be a member of the community to like this thread")
		}
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

func (s *LikeService) IsLiked(userID, threadID uint) bool {
	ok, _ := s.likeRepo.Exists(userID, threadID)
	return ok
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
