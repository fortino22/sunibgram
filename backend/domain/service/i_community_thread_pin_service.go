package service

import (
	"context"

	communitythreadpinmodel "github.com/sunibgram/domain/model/community_thread_pin"
)

type ICommunityThreadPinService interface {
	Pin(ctx context.Context, communityID, actorUserID, threadID uint) error
	Unpin(ctx context.Context, communityID, actorUserID, threadID uint) error
	GetPinned(communityID uint) (*communitythreadpinmodel.CommunityThreadPin, error)
}
