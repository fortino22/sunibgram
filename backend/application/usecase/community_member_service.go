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
}

func isModeratorOrAbove(role model.CommunityMemberRole) bool {
	return role == model.CommunityMemberRoleModerator ||
		role == model.CommunityMemberRoleAdmin ||
		role == model.CommunityMemberRoleOwner
}

func isAdminOrAbove(role model.CommunityMemberRole) bool {
	return role == model.CommunityMemberRoleAdmin || role == model.CommunityMemberRoleOwner
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
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.memberRepo.FindByCommunityID(communityID, page, limit)
}

func (s *CommunityMemberService) RemoveMember(ctx context.Context, communityID, actorUserID, targetUserID uint) error {
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !isModeratorOrAbove(actor.Role) {
		return errors.New("forbidden")
	}
	target, err := s.memberRepo.FindMember(communityID, targetUserID)
	if err != nil {
		return errors.New("member not found")
	}
	if target.Role == model.CommunityMemberRoleOwner {
		return errors.New("cannot remove the owner")
	}
	// moderator can only remove members, not admins/other moderators
	if actor.Role == model.CommunityMemberRoleModerator && target.Role != model.CommunityMemberRoleMember {
		return errors.New("forbidden")
	}
	if err := s.memberRepo.Delete(communityID, targetUserID); err != nil {
		return err
	}
	return s.memberRepo.UpdateMemberCount(communityID, -1)
}

func (s *CommunityMemberService) IsMember(communityID, userID uint) bool {
	_, err := s.memberRepo.FindMember(communityID, userID)
	return err == nil
}

func (s *CommunityMemberService) UpdateRole(ctx context.Context, communityID, actorUserID, targetUserID uint, role model.CommunityMemberRole) error {
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !isAdminOrAbove(actor.Role) {
		return errors.New("forbidden")
	}
	target, err := s.memberRepo.FindMember(communityID, targetUserID)
	if err != nil {
		return errors.New("member not found")
	}
	if target.Role == model.CommunityMemberRoleOwner {
		return errors.New("cannot change owner role")
	}
	// admin can only assign up to moderator; owner can assign up to admin
	if actor.Role == model.CommunityMemberRoleAdmin && role == model.CommunityMemberRoleAdmin {
		return errors.New("forbidden")
	}
	if role == model.CommunityMemberRoleOwner {
		return errors.New("cannot assign owner role")
	}
	return s.memberRepo.UpdateRole(communityID, targetUserID, role)
}
