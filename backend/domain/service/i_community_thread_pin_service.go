package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type ICommunityThreadPinService interface {
	Pin(ctx context.Context, communityID, actorUserID, threadID uint) error
	Unpin(ctx context.Context, communityID, actorUserID, threadID uint) error
	GetPinned(communityID uint) (*model.CommunityThreadPin, error)
}
