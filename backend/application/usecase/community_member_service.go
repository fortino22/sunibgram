package usecase

import (
	"context"
	"errors"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type CommunityMemberService struct {
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
}

func NewCommunityMemberService(
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
) service.ICommunityMemberService {
	return &CommunityMemberService{communityRepo: communityRepo, memberRepo: memberRepo}
}

func (s *CommunityMemberService) Join(ctx context.Context, communityID, userID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if community.Visibility == model.CommunityVisibilityPrivate {
		return errors.New("community is private, send a join request instead")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err == nil {
		return errors.New("already a member")
	}
	member := &model.CommunityMember{
		CommunityID: communityID,
		UserID:      userID,
		Role:        model.CommunityMemberRoleMember,
	}
	if err := s.memberRepo.Save(member); err != nil {
		return err
	}
	return s.memberRepo.UpdateMemberCount(communityID, 1)
}

func (s *CommunityMemberService) Leave(ctx context.Context, communityID, userID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if community.UserID == userID {
		return errors.New("owner cannot leave the community")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return errors.New("not a member")
	}
	if err := s.memberRepo.Delete(communityID, userID); err != nil {
		return err
	}
	return s.memberRepo.UpdateMemberCount(communityID, -1)
}

func (s *CommunityMemberService) GetMembers(communityID uint, page, limit int) ([]model.CommunityMember, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.memberRepo.FindByCommunityID(communityID, page, limit)
}
