package communityrequestuc

import (
	"context"
	"errors"

	"github.com/sunibgram/application/helper"
	communityuc "github.com/sunibgram/application/usecase/community"
	communitymodel "github.com/sunibgram/domain/model/community"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communityrequestmodel "github.com/sunibgram/domain/model/community_request"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	communityrequestrepoiface "github.com/sunibgram/domain/repository/community_request"
	communityrequestsvciface "github.com/sunibgram/domain/service/community_request"
)

type CommunityRequestService struct {
	communityRepo communityrepoiface.ICommunityRepository
	memberRepo    communitymemberrepoiface.ICommunityMemberRepository
	requestRepo   communityrequestrepoiface.ICommunityRequestRepository
	cacheRepo     cacherepoiface.ICacheRepository
}

func NewCommunityRequestService(
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	requestRepo communityrequestrepoiface.ICommunityRequestRepository,
	cacheRepo cacherepoiface.ICacheRepository,
) communityrequestsvciface.ICommunityRequestService {
	return &CommunityRequestService{
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		requestRepo:   requestRepo,
		cacheRepo:     cacheRepo,
	}
}

func (s *CommunityRequestService) Send(ctx context.Context, communityID, userID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	if community.Visibility != communitymodel.CommunityVisibilityPrivate {
		return errors.New("community is not private, use join instead")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err == nil {
		return errors.New("already a member")
	}
	if existing, err := s.requestRepo.FindRequest(communityID, userID); err == nil {
		switch existing.Status {
		case communityrequestmodel.CommunityRequestStatusPending:
			return errors.New("request already pending")
		case communityrequestmodel.CommunityRequestStatusApproved:
			return errors.New("already a member")
		case communityrequestmodel.CommunityRequestStatusDeclined:
			return s.requestRepo.UpdateStatus(communityID, userID, communityrequestmodel.CommunityRequestStatusPending)
		}
	}
	req := &communityrequestmodel.CommunityRequest{
		CommunityID: communityID,
		UserID:      userID,
		Status:      communityrequestmodel.CommunityRequestStatusPending,
	}
	return s.requestRepo.Save(req)
}

func (s *CommunityRequestService) Cancel(ctx context.Context, communityID, userID uint) error {
	req, err := s.requestRepo.FindRequest(communityID, userID)
	if err != nil {
		return errors.New("request not found")
	}
	if req.Status != communityrequestmodel.CommunityRequestStatusPending {
		return errors.New("request is not pending")
	}
	return s.requestRepo.Delete(communityID, userID)
}

func (s *CommunityRequestService) Approve(ctx context.Context, communityID, adminUserID, targetUserID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	adminMember, err := s.memberRepo.FindMember(communityID, adminUserID)
	if err != nil || !communityuc.IsCommunityAdmin(adminMember.Role) {
		return errors.New("forbidden")
	}
	req, err := s.requestRepo.FindRequest(communityID, targetUserID)
	if err != nil {
		return errors.New("join request does not exist")
	}
	if req.Status != communityrequestmodel.CommunityRequestStatusPending {
		return errors.New("join request is already " + string(req.Status))
	}
	if err := s.requestRepo.UpdateStatus(communityID, targetUserID, communityrequestmodel.CommunityRequestStatusApproved); err != nil {
		return err
	}
	member := &communitymembermodel.CommunityMember{
		CommunityID: communityID,
		UserID:      targetUserID,
		Role:        communitymembermodel.CommunityMemberRoleMember,
	}
	if err := s.memberRepo.Save(member); err != nil {
		return err
	}
	communityuc.InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if err := s.memberRepo.UpdateMemberCount(communityID, 1); err != nil {
		return err
	}
	return nil
}

func (s *CommunityRequestService) Decline(ctx context.Context, communityID, adminUserID, targetUserID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	adminMember, err := s.memberRepo.FindMember(communityID, adminUserID)
	if err != nil || !communityuc.IsCommunityAdmin(adminMember.Role) {
		return errors.New("forbidden")
	}
	req, err := s.requestRepo.FindRequest(communityID, targetUserID)
	if err != nil {
		return errors.New("join request does not exist")
	}
	if req.Status != communityrequestmodel.CommunityRequestStatusPending {
		return errors.New("join request is already " + string(req.Status))
	}
	return s.requestRepo.UpdateStatus(communityID, targetUserID, communityrequestmodel.CommunityRequestStatusDeclined)
}

func (s *CommunityRequestService) GetPending(ctx context.Context, communityID, adminUserID uint, page, limit int) ([]communityrequestmodel.CommunityRequest, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	adminMember, err := s.memberRepo.FindMember(communityID, adminUserID)
	if err != nil || !communityuc.IsCommunityAdmin(adminMember.Role) {
		return nil, 0, errors.New("forbidden")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.requestRepo.FindPending(communityID, page, limit)
}
