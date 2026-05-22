package repository

import (
	communitymodel "github.com/sunibgram/domain/model/community"
	communitypinmodel "github.com/sunibgram/domain/model/community_pin"
)

type ICommunityPinRepository interface {
	Pin(pin *communitypinmodel.UserCommunityPin) error
	Unpin(userID, communityID uint) error
	Exists(userID, communityID uint) (bool, error)
	CountByUser(userID uint) (int64, error)
	FindMyPins(userID uint, page, limit int) ([]communitymodel.Community, int64, error)
	InvalidateUser(userID uint)
	InvalidateByCommunity(communityID uint)
}
