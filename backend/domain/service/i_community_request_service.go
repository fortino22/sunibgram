package service

import (
	"context"

	communityrequestmodel "github.com/sunibgram/domain/model/community_request"
)

type ICommunityRequestService interface {
	Send(ctx context.Context, communityID, userID uint) error
	Cancel(ctx context.Context, communityID, userID uint) error
	Approve(ctx context.Context, communityID, adminUserID, targetUserID uint) error
	Decline(ctx context.Context, communityID, adminUserID, targetUserID uint) error
	GetPending(ctx context.Context, communityID, adminUserID uint, page, limit int) ([]communityrequestmodel.CommunityRequest, int64, error)
}
