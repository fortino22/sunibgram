package usecase

import (
	"errors"

	"github.com/sunibgram/application/helper"
	communitymodel "github.com/sunibgram/domain/model/community"
	communitypinmodel "github.com/sunibgram/domain/model/community_pin"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	communitypinrepoiface "github.com/sunibgram/domain/repository/community_pin"
	communitypinsvciface "github.com/sunibgram/domain/service/community_pin"
)

type CommunityPinService struct {
	pinRepo       communitypinrepoiface.ICommunityPinRepository
	communityRepo communityrepoiface.ICommunityRepository
	memberRepo    communitymemberrepoiface.ICommunityMemberRepository
}

func NewCommunityPinService(
	pinRepo communitypinrepoiface.ICommunityPinRepository,
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
) communitypinsvciface.ICommunityPinService {
	return &CommunityPinService{pinRepo: pinRepo, communityRepo: communityRepo, memberRepo: memberRepo}
}

func (s *CommunityPinService) Pin(userID, communityID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if community.Visibility == communitymodel.CommunityVisibilityPrivate {
		if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
			return errors.New("you must be a member to pin this community")
		}
	}
	exists, err := s.pinRepo.Exists(userID, communityID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already pinned")
	}
	count, err := s.pinRepo.CountByUser(userID)
	if err != nil {
		return err
	}
	if count >= maxCommunityPinsPerUser {
		return errors.New("maximum of 3 pinned communities reached")
	}
	return s.pinRepo.Pin(&communitypinmodel.UserCommunityPin{UserID: userID, CommunityID: communityID})
}

const maxCommunityPinsPerUser = 3

func (s *CommunityPinService) Unpin(userID, communityID uint) error {
	exists, err := s.pinRepo.Exists(userID, communityID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not pinned")
	}
	return s.pinRepo.Unpin(userID, communityID)
}

func (s *CommunityPinService) GetMyPins(userID uint, page, limit int) ([]communitymodel.Community, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.pinRepo.FindMyPins(userID, page, limit)
}
