package usecase

import (
	"context"
	"errors"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type CommunityMemberService struct {
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	cacheRepo     repository.ICacheRepository
}

func isCommunityAdmin(role model.CommunityMemberRole) bool {
	return role == model.CommunityMemberRoleAdmin
}

func NewCommunityMemberService(
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	cacheRepo repository.ICacheRepository,
) service.ICommunityMemberService {
	return &CommunityMemberService{communityRepo: communityRepo, memberRepo: memberRepo, cacheRepo: cacheRepo}
}

func (s *CommunityMemberService) Join(ctx context.Context, communityID, userID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err == nil {
		return errors.New("already a member")
	}
	if community.Visibility == model.CommunityVisibilityPrivate {
		return errors.New("community is private, send a join request instead")
	}
	member := &model.CommunityMember{
		CommunityID: communityID,
		UserID:      userID,
		Role:        model.CommunityMemberRoleMember,
	}
	if err := s.memberRepo.Save(member); err != nil {
		return err
	}
	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if err := s.memberRepo.UpdateMemberCount(communityID, 1); err != nil {
		return err
	}
	return nil
}

func (s *CommunityMemberService) Leave(ctx context.Context, communityID, userID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if community.UserID == userID {
		return errors.New("creator cannot leave the community")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return errors.New("not a member")
	}
	if err := s.memberRepo.Delete(communityID, userID); err != nil {
		return err
	}
	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if err := s.memberRepo.UpdateMemberCount(communityID, -1); err != nil {
		return err
	}
	return nil
}

func (s *CommunityMemberService) GetMembers(communityID uint, page, limit int) ([]model.CommunityMember, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.memberRepo.FindByCommunityID(communityID, page, limit)
}

func (s *CommunityMemberService) RemoveMember(ctx context.Context, communityID, actorUserID, targetUserID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !isCommunityAdmin(actor.Role) {
		return errors.New("forbidden")
	}
	if community.UserID == targetUserID {
		return errors.New("cannot remove the community creator")
	}
	if _, err := s.memberRepo.FindMember(communityID, targetUserID); err != nil {
		return errors.New("member not found")
	}
	if err := s.memberRepo.Delete(communityID, targetUserID); err != nil {
		return err
	}
	invalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if err := s.memberRepo.UpdateMemberCount(communityID, -1); err != nil {
		return err
	}
	return nil
}

func (s *CommunityMemberService) IsMember(communityID, userID uint) bool {
	_, err := s.memberRepo.FindMember(communityID, userID)
	return err == nil
}

func (s *CommunityMemberService) UpdateRole(ctx context.Context, communityID, actorUserID, targetUserID uint, role model.CommunityMemberRole) error {
	if role != model.CommunityMemberRoleMember && role != model.CommunityMemberRoleAdmin {
		return errors.New("invalid role")
	}
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !isCommunityAdmin(actor.Role) {
		return errors.New("forbidden")
	}
	if community.UserID == targetUserID {
		return errors.New("cannot change the community creator role")
	}
	if _, err := s.memberRepo.FindMember(communityID, targetUserID); err != nil {
		return errors.New("member not found")
	}
	return s.memberRepo.UpdateRole(communityID, targetUserID, role)
}
