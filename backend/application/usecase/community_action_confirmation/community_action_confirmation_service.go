package communityactionconfirmationuc

import (
	"context"
	"time"

	communityactionconfirmationmodel "github.com/sunibgram/domain/model/community_action_confirmation"
	communityactionconfirmationrepoiface "github.com/sunibgram/domain/repository/community_action_confirmation"
	communityactionconfirmationsvciface "github.com/sunibgram/domain/service/community_action_confirmation"
)

type CommunityActionConfirmationService struct {
	repo communityactionconfirmationrepoiface.ICommunityActionConfirmationRepository
}

func NewCommunityActionConfirmationService(repo communityactionconfirmationrepoiface.ICommunityActionConfirmationRepository) communityactionconfirmationsvciface.ICommunityActionConfirmationService {
	return &CommunityActionConfirmationService{repo: repo}
}

func (s *CommunityActionConfirmationService) RecordAdminTransfer(_ context.Context, communityID, fromUserID, toUserID uint) error {
	target := toUserID
	rec := &communityactionconfirmationmodel.CommunityActionConfirmation{
		CommunityID:  communityID,
		ActorUserID:  fromUserID,
		TargetUserID: &target,
		Type:         communityactionconfirmationmodel.ActionAdminTransfer,
		CreatedAt:    time.Now(),
	}
	return s.repo.Save(rec)
}

func (s *CommunityActionConfirmationService) RecordCommunityDeletion(_ context.Context, communityID, actorUserID uint, communityName string) error {
	rec := &communityactionconfirmationmodel.CommunityActionConfirmation{
		CommunityID:   communityID,
		ActorUserID:   actorUserID,
		Type:          communityactionconfirmationmodel.ActionDeleteCommunity,
		CommunityName: communityName,
		CreatedAt:     time.Now(),
	}
	return s.repo.Save(rec)
}
