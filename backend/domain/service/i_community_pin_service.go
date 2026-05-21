package service

import "github.com/sunibgram/domain/model"

type ICommunityPinService interface {
	Pin(userID, communityID uint) error
	Unpin(userID, communityID uint) error
	GetMyPins(userID uint, page, limit int) ([]model.Community, int64, error)
}
