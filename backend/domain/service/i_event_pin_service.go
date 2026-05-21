package service

import "github.com/sunibgram/domain/model"

type IEventPinService interface {
	Pin(userID uint, eventID string) error
	Unpin(userID uint, eventID string) error
	GetMyPins(userID uint, page, limit int) ([]model.Event, int64, error)
}
