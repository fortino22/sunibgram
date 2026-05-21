package repository

import "github.com/sunibgram/domain/model"

type IEventThreadPinRepository interface {
	Pin(pin *model.EventThreadPin) error
	Unpin(eventID string, threadID uint) error
	Exists(eventID string, threadID uint) (bool, error)
	FindByEventID(eventID string) (*model.EventThreadPin, error)
}
