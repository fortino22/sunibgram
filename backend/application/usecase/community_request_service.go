package usecase

import (
	"context"
	"errors"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type CommunityRequestService struct {
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	requestRepo   repository.ICommunityRequestRepository
	cacheRepo     repository.ICacheRepository
}

func NewCommunityRequestService(
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	requestRepo repository.ICommunityRequestRepository,
	cacheRepo repository.ICacheRepository,
) service.ICommunityRequestService {
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
	if community.Visibility != model.CommunityVisibilityPrivate {
		return errors.New("community is not private, use join instead")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err == nil {
		return errors.New("already a member")
	}
	if req, err := s.requestRepo.FindRequest(communityID, userID); err == nil && req.Status == model.CommunityRequestStatusPending {
		return errors.New("request already pending")
	}
	req := &model.CommunityRequest{
		CommunityID: communityID,
		UserID:      userID,
		Status:      model.CommunityRequestStatusPending,
	}
	return s.requestRepo.Save(req)
}

func (s *CommunityRequestService) Cancel(ctx context.Context, communityID, userID uint) error {
	req, err := s.requestRepo.FindRequest(communityID, userID)
	if err != nil {
		return errors.New("request not found")
	}
	if req.Status != model.CommunityRequestStatusPending {
		return errors.New("request is not pending")
	}
	return s.requestRepo.Delete(communityID, userID)
}

func (s *CommunityRequestService) Approve(ctx context.Context, communityID, adminUserID, targetUserID uint) error {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return errors.New("community not found")
	}
	adminMember, err := s.memberRepo.FindMember(communityID, adminUserID)
	if err != nil || !isModeratorOrAbove(adminMember.Role) {
		return errors.New("forbidden")
	}
	req, err := s.requestRepo.FindRequest(communityID, targetUserID)
	if err != nil {
		return errors.New("join request does not exist")
	}
	if req.Status != model.CommunityRequestStatusPending {
		return errors.New("join request is already " + string(req.Status))
	}
	if err := s.requestRepo.UpdateStatus(communityID, targetUserID, model.CommunityRequestStatusApproved); err != nil {
		return err
	}
	member := &model.CommunityMember{
		CommunityID: communityID,
		UserID:      targetUserID,
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

func (s *CommunityRequestService) Decline(ctx context.Context, communityID, adminUserID, targetUserID uint) error {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return errors.New("community not found")
	}
	adminMember, err := s.memberRepo.FindMember(communityID, adminUserID)
	if err != nil || !isModeratorOrAbove(adminMember.Role) {
		return errors.New("forbidden")
	}
	req, err := s.requestRepo.FindRequest(communityID, targetUserID)
	if err != nil {
		return errors.New("join request does not exist")
	}
	if req.Status != model.CommunityRequestStatusPending {
		return errors.New("join request is already " + string(req.Status))
	}
	return s.requestRepo.UpdateStatus(communityID, targetUserID, model.CommunityRequestStatusDeclined)
}

func (s *CommunityRequestService) GetPending(ctx context.Context, communityID, adminUserID uint, page, limit int) ([]model.CommunityRequest, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	adminMember, err := s.memberRepo.FindMember(communityID, adminUserID)
	if err != nil || !isModeratorOrAbove(adminMember.Role) {
		return nil, 0, errors.New("forbidden")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.requestRepo.FindPending(communityID, page, limit)
}
