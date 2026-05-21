package repository

import "github.com/sunibgram/domain/model"

type ICommunityThreadPinRepository interface {
	Pin(pin *model.CommunityThreadPin) error
	Unpin(communityID, threadID uint) error
	Exists(communityID, threadID uint) (bool, error)
	FindByCommunityID(communityID uint) (*model.CommunityThreadPin, error)
}
