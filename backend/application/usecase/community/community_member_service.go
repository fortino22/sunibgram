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
	communitymembersvciface "github.com/sunibgram/domain/service/community_member"
)

type CommunityMemberService struct {
	communityRepo communityrepoiface.ICommunityRepository
	memberRepo    communitymemberrepoiface.ICommunityMemberRepository
	followRepo    followrepoiface.IFollowRepository
	cacheRepo     cacherepoiface.ICacheRepository
}

func IsCommunityAdmin(role communitymembermodel.CommunityMemberRole) bool {
	return role == communitymembermodel.CommunityMemberRoleAdmin
}

func NewCommunityMemberService(
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	followRepo followrepoiface.IFollowRepository,
	cacheRepo cacherepoiface.ICacheRepository,
) communitymembersvciface.ICommunityMemberService {
	return &CommunityMemberService{
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		followRepo:    followRepo,
		cacheRepo:     cacheRepo,
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
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return errors.New("not a member")
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

func (s *CommunityMemberService) GetMembers(communityID uint, page, limit int) ([]communitymembermodel.CommunityMember, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.memberRepo.FindByCommunityID(communityID, page, limit)
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

func (s *CommunityMemberService) RemoveMember(ctx context.Context, communityID, actorUserID, targetUserID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
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
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !IsCommunityAdmin(actor.Role) {
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
