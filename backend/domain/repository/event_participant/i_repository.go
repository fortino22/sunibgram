package eventparticipantrepoiface

import (
	eventparticipantmodel "github.com/sunibgram/domain/model/event_participant"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IEventParticipantRepository interface {
	Join(p *eventparticipantmodel.EventParticipant) error
	Leave(userID uint, eventID string) error
	Exists(userID uint, eventID string) (bool, error)
	FindMyJoined(userID uint, page, limit int) ([]threadmodel.Event, int64, error)
	InvalidateUser(userID uint)
	InvalidateByEvent(eventID string)
}
