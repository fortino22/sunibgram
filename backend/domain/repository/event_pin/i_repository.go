package eventpinrepoiface

import (
	eventpinmodel "github.com/sunibgram/domain/model/event_pin"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IEventPinRepository interface {
	Pin(pin *eventpinmodel.UserEventPin) error
	Unpin(userID uint, eventID string) error
	Exists(userID uint, eventID string) (bool, error)
	CountByUser(userID uint) (int64, error)
	FindMyPins(userID uint, page, limit int) ([]threadmodel.Event, int64, error)
	InvalidateUser(userID uint)
	InvalidateByEvent(eventID string)
}
