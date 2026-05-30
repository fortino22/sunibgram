package communityuc

import (
	"context"
	"errors"

	"github.com/sunibgram/application/helper"
	communitymodel "github.com/sunibgram/domain/model/community"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	communityactionconfirmationsvciface "github.com/sunibgram/domain/service/community_action_confirmation"
	communitymembersvciface "github.com/sunibgram/domain/service/community_member"
)

type CommunityMemberService struct {
	communityRepo communityrepoiface.ICommunityRepository
	memberRepo    communitymemberrepoiface.ICommunityMemberRepository
	followRepo    followrepoiface.IFollowRepository
	blockRepo     userblockrepoiface.IUserBlockRepository
	cacheRepo     cacherepoiface.ICacheRepository
	auditSvc      communityactionconfirmationsvciface.ICommunityActionConfirmationService
}

func IsCommunityAdmin(role communitymembermodel.CommunityMemberRole) bool {
	return role == communitymembermodel.CommunityMemberRoleAdmin
}

func NewCommunityMemberService(
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	followRepo followrepoiface.IFollowRepository,
	blockRepo userblockrepoiface.IUserBlockRepository,
	cacheRepo cacherepoiface.ICacheRepository,
	auditSvc communityactionconfirmationsvciface.ICommunityActionConfirmationService,
) communitymembersvciface.ICommunityMemberService {
	return &CommunityMemberService{
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		followRepo:    followRepo,
		blockRepo:     blockRepo,
		cacheRepo:     cacheRepo,
		auditSvc:      auditSvc,
	}
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
	if community.Visibility == communitymodel.CommunityVisibilityPrivate {
		return errors.New("community is private, send a join request instead")
	}
	member := &communitymembermodel.CommunityMember{
		CommunityID: communityID,
		UserID:      userID,
		Role:        communitymembermodel.CommunityMemberRoleMember,
	}
	if err := s.memberRepo.Save(member); err != nil {
		return err
	}
	InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
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
	membership, err := s.memberRepo.FindMember(communityID, userID)
	if err != nil {
		return errors.New("not a member")
	}
	if IsCommunityAdmin(membership.Role) {
		adminCount, err := s.memberRepo.CountAdmins(communityID)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return errors.New("sole admin cannot leave; transfer admin or delete community")
		}
	}
	if err := s.memberRepo.Delete(communityID, userID); err != nil {
		return err
	}
	InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if err := s.memberRepo.UpdateMemberCount(communityID, -1); err != nil {
		return err
	}
	return nil
}

func (s *CommunityMemberService) GetMembers(viewerID, communityID uint, page, limit int) ([]communitymembermodel.CommunityMember, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	members, total, err := s.memberRepo.FindByCommunityID(communityID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return s.filterBlockedMembers(viewerID, members), total, nil
}

func (s *CommunityMemberService) GetCommunityUsers(ctx context.Context, communityID, viewerID uint, role *communitymembermodel.CommunityMemberRole, query string, page, limit int) ([]communitymembersvciface.CommunityUserItem, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)

	members, total, err := s.memberRepo.FindCommunityUsers(communityID, role, query, page, limit)
	if err != nil {
		return nil, 0, err
	}
	members = s.filterBlockedMembers(viewerID, members)

	userIDs := make([]uint, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	followSet, err := s.followRepo.GetFollowingSet(ctx, viewerID, userIDs)
	if err != nil {
		return nil, 0, err
	}

	items := make([]communitymembersvciface.CommunityUserItem, len(members))
	for i, m := range members {
		avatarObject := ""
		if m.Profile != nil {
			avatarObject = m.Profile.AvatarObjectName
		}
		items[i] = communitymembersvciface.CommunityUserItem{
			UserID:           m.UserID,
			Username:         m.User.UserName,
			Name:             m.User.Name,
			AvatarObjectName: avatarObject,
			IsPremium:        m.User.IsPremium,
			Role:             m.Role,
			IsFollowing:      followSet[m.UserID],
		}
	}
	return items, total, nil
}

// filterBlockedMembers drops members in a mutual block relationship with the
// viewer (mutual hide). Page-bounded (≤20) so the per-user IsBlocked loop is
// cheap. Filtered at the service layer; total stays approximate.
func (s *CommunityMemberService) filterBlockedMembers(viewerID uint, members []communitymembermodel.CommunityMember) []communitymembermodel.CommunityMember {
	if s.blockRepo == nil || viewerID == 0 || len(members) == 0 {
		return members
	}
	out := make([]communitymembermodel.CommunityMember, 0, len(members))
	for _, m := range members {
		if m.UserID == viewerID {
			out = append(out, m)
			continue
		}
		blocked, err := s.blockRepo.IsBlocked(viewerID, m.UserID)
		if err == nil && blocked {
			continue
		}
		out = append(out, m)
	}
	return out
}

func (s *CommunityMemberService) RemoveMember(ctx context.Context, communityID, actorUserID, targetUserID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !IsCommunityAdmin(actor.Role) {
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
	InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if err := s.memberRepo.UpdateMemberCount(communityID, -1); err != nil {
		return err
	}
	return nil
}

func (s *CommunityMemberService) IsMember(communityID, userID uint) bool {
	_, err := s.memberRepo.FindMember(communityID, userID)
	return err == nil
}

func (s *CommunityMemberService) UpdateRole(ctx context.Context, communityID, actorUserID, targetUserID uint, role communitymembermodel.CommunityMemberRole) error {
	if role != communitymembermodel.CommunityMemberRoleMember && role != communitymembermodel.CommunityMemberRoleAdmin {
		return errors.New("invalid role")
	}
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !IsCommunityAdmin(actor.Role) {
		return errors.New("forbidden")
	}
	target, err := s.memberRepo.FindMember(communityID, targetUserID)
	if err != nil {
		return errors.New("member not found")
	}

	if role == communitymembermodel.CommunityMemberRoleAdmin {
		if targetUserID == actorUserID {
			return nil
		}
		if target.Role == communitymembermodel.CommunityMemberRoleAdmin {
			return nil
		}
		if err := s.memberRepo.TransferAdmin(communityID, actorUserID, targetUserID); err != nil {
			return err
		}
		if s.auditSvc != nil {
			_ = s.auditSvc.RecordAdminTransfer(ctx, communityID, actorUserID, targetUserID)
		}
		return nil
	}

	if targetUserID == actorUserID {
		count, err := s.memberRepo.CountAdmins(communityID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return errors.New("sole admin cannot self-demote; transfer admin or delete community")
		}
	}
	return s.memberRepo.UpdateRole(communityID, targetUserID, role)
}
