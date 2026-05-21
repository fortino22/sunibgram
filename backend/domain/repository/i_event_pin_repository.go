package repository

import "github.com/sunibgram/domain/model"

type IEventPinRepository interface {
	Pin(pin *model.UserEventPin) error
	Unpin(userID uint, eventID string) error
	Exists(userID uint, eventID string) (bool, error)
	FindMyPins(userID uint, page, limit int) ([]model.Event, int64, error)
	InvalidateUser(userID uint)
}
