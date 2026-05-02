package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type ICommunityRequestService interface {
	Send(ctx context.Context, communityID, userID uint) error
	Cancel(ctx context.Context, communityID, userID uint) error
	Approve(ctx context.Context, communityID, adminUserID, targetUserID uint) error
	Decline(ctx context.Context, communityID, adminUserID, targetUserID uint) error
	GetPending(ctx context.Context, communityID, adminUserID uint, page, limit int) ([]model.CommunityRequest, int64, error)
}
