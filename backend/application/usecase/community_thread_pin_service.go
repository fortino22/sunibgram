package usecase

import (
	"context"
	"errors"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type CommunityThreadPinService struct {
	pinRepo       repository.ICommunityThreadPinRepository
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	threadRepo    repository.IThreadRepository
}

func NewCommunityThreadPinService(
	pinRepo repository.ICommunityThreadPinRepository,
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	threadRepo repository.IThreadRepository,
) service.ICommunityThreadPinService {
	return &CommunityThreadPinService{
		pinRepo:       pinRepo,
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		threadRepo:    threadRepo,
	}
}

func (s *CommunityThreadPinService) Pin(ctx context.Context, communityID, actorUserID, threadID uint) error {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return errors.New("community not found")
	}
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !isCommunityAdmin(actor.Role) {
		return errors.New("forbidden")
	}
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.CommunityID == nil || *thread.CommunityID != communityID {
		return errors.New("thread does not belong to this community")
	}

	existing, err := s.pinRepo.FindByCommunityID(communityID)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.ThreadID == threadID {
			return errors.New("already pinned")
		}
		if err := s.pinRepo.Unpin(communityID, existing.ThreadID); err != nil {
			return err
		}
	}

	return s.pinRepo.Pin(&model.CommunityThreadPin{CommunityID: communityID, ThreadID: threadID})
}

func (s *CommunityThreadPinService) Unpin(ctx context.Context, communityID, actorUserID, threadID uint) error {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return errors.New("community not found")
	}
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !isCommunityAdmin(actor.Role) {
		return errors.New("forbidden")
	}
	exists, err := s.pinRepo.Exists(communityID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not pinned")
	}
	return s.pinRepo.Unpin(communityID, threadID)
}

func (s *CommunityThreadPinService) GetPinned(communityID uint) (*model.CommunityThreadPin, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, errors.New("community not found")
	}
	return s.pinRepo.FindByCommunityID(communityID)
}
