package usecase

import (
	"errors"
	"time"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type RepostService struct {
	repostRepo   repository.IRepostRepository
	threadRepo   repository.IThreadRepository
	notification service.INotificationService
}

func NewRepostService(
	repostRepo repository.IRepostRepository,
	threadRepo repository.IThreadRepository,
	notification service.INotificationService,
) service.IRepostService {
	return &RepostService{repostRepo: repostRepo, threadRepo: threadRepo, notification: notification}
}

func (s *RepostService) Repost(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
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
	s.threadRepo.InvalidateThread(threadID)
	helper.SendInteractionNotification(s.notification, "repost", "reposted your thread", userID, thread.UserID, threadID)
	return nil
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
	s.threadRepo.InvalidateThread(threadID)
	return nil
}

func (s *RepostService) GetMyReposts(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.repostRepo.FindByUserID(userID, page, limit)
}
