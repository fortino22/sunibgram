package repostuc

import (
	"errors"
	"time"

	"github.com/sunibgram/application/helper"
	repostmodel "github.com/sunibgram/domain/model/repost"
	threadmodel "github.com/sunibgram/domain/model/thread"
	repostrepoiface "github.com/sunibgram/domain/repository/repost"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
	repostsvciface "github.com/sunibgram/domain/service/repost"
)

type RepostService struct {
	repostRepo   repostrepoiface.IRepostRepository
	threadRepo   threadrepoiface.IThreadRepository
	notification notificationsvciface.INotificationService
}

func NewRepostService(
	repostRepo repostrepoiface.IRepostRepository,
	threadRepo threadrepoiface.IThreadRepository,
	notification notificationsvciface.INotificationService,
) repostsvciface.IRepostService {
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
	if err := s.repostRepo.Save(&repostmodel.Repost{UserID: userID, ThreadID: threadID}); err != nil {
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

func (s *RepostService) AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return s.repostRepo.AreThreadsReposted(userID, threadIDs)
}

func (s *RepostService) GetMyReposts(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.repostRepo.FindByUserID(userID, page, limit)
}
