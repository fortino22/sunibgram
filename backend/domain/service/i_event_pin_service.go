package service

import threadmodel "github.com/sunibgram/domain/model/thread"

type IEventPinService interface {
	Pin(userID uint, eventID string) error
	Unpin(userID uint, eventID string) error
	GetMyPins(userID uint, page, limit int) ([]threadmodel.Event, int64, error)
}
