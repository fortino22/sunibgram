package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type LikeService struct {
	likeRepo      repository.ILikeRepository
	threadRepo    repository.IThreadRepository
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	settingsRepo  repository.IUserSettingsRepository
	followRepo    repository.IFollowRepository
	notification  service.INotificationService
}

func NewLikeService(
	likeRepo repository.ILikeRepository,
	threadRepo repository.IThreadRepository,
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	settingsRepo repository.IUserSettingsRepository,
	followRepo repository.IFollowRepository,
	notification service.INotificationService,
) service.ILikeService {
	return &LikeService{
		likeRepo:      likeRepo,
		threadRepo:    threadRepo,
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		settingsRepo:  settingsRepo,
		followRepo:    followRepo,
		notification:  notification,
	}
}

func (s *LikeService) Like(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if err := canViewPrivateUser(context.Background(), s.settingsRepo, s.followRepo, &userID, thread.UserID); err != nil {
		return err
	}
	if thread.CommunityID != nil {
		community, err := s.communityRepo.FindByID(*thread.CommunityID)
		if err != nil {
			return errors.New("community not found")
		}
		if community.Visibility == model.CommunityVisibilityPrivate {
			if _, err := s.memberRepo.FindMember(*thread.CommunityID, userID); err != nil {
				return errors.New("you must be a member of the community to like this thread")
			}
		}
	}
	exists, err := s.likeRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already liked")
	}
	if err := s.likeRepo.Save(&model.Like{UserID: userID, ThreadID: threadID}); err != nil {
		return err
	}
	s.threadRepo.InvalidateThread(threadID)
	helper.SendInteractionNotification(s.notification, "like", "liked your thread", userID, thread.UserID, threadID)
	return nil
}

func (s *LikeService) Unlike(userID, threadID uint) error {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if err := canViewPrivateUser(context.Background(), s.settingsRepo, s.followRepo, &userID, thread.UserID); err != nil {
		return err
	}
	exists, err := s.likeRepo.Exists(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not liked")
	}
	if err := s.likeRepo.Delete(userID, threadID); err != nil {
		return err
	}
	s.threadRepo.InvalidateThread(threadID)
	return nil
}

func (s *LikeService) IsLiked(userID, threadID uint) bool {
	ok, _ := s.likeRepo.Exists(userID, threadID)
	return ok
}

func (s *LikeService) AreThreadsLiked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return s.likeRepo.AreThreadsLiked(userID, threadIDs)
}

func (s *LikeService) GetMyLikes(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.likeRepo.FindByUserID(userID, page, limit)
}
