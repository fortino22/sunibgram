package eventthreadpinrepoiface

import eventthreadpinmodel "github.com/sunibgram/domain/model/event_thread_pin"

type IEventThreadPinRepository interface {
	Pin(pin *eventthreadpinmodel.EventThreadPin) error
	Unpin(eventID string, threadID uint) error
	Exists(eventID string, threadID uint) (bool, error)
	FindByEventID(eventID string) (*eventthreadpinmodel.EventThreadPin, error)
}
