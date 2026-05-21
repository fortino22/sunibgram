package repository

import "github.com/sunibgram/domain/model"

type ICommunityPinRepository interface {
	Pin(pin *model.UserCommunityPin) error
	Unpin(userID, communityID uint) error
	Exists(userID, communityID uint) (bool, error)
	FindMyPins(userID uint, page, limit int) ([]model.Community, int64, error)
	InvalidateUser(userID uint)
}
