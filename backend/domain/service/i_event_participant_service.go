package service

import "github.com/sunibgram/domain/model"

type IEventParticipantService interface {
	Join(userID uint, eventID string) error
	Leave(userID uint, eventID string) error
	GetMyJoined(userID uint, page, limit int) ([]model.Event, int64, error)
}
