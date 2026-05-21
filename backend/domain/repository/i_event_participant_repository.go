package repository

import "github.com/sunibgram/domain/model"

type IEventParticipantRepository interface {
	Join(p *model.EventParticipant) error
	Leave(userID uint, eventID string) error
	Exists(userID uint, eventID string) (bool, error)
	FindMyJoined(userID uint, page, limit int) ([]model.Event, int64, error)
	InvalidateUser(userID uint)
}
