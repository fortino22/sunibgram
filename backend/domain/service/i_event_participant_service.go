package service

import threadmodel "github.com/sunibgram/domain/model/thread"

type IEventParticipantService interface {
	Join(userID uint, eventID string) error
	Leave(userID uint, eventID string) error
	GetMyJoined(userID uint, page, limit int) ([]threadmodel.Event, int64, error)
}
