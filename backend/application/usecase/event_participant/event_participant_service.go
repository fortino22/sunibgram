package eventparticipantuc

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/sunibgram/application/helper"
	eventparticipantmodel "github.com/sunibgram/domain/model/event_participant"
	threadmodel "github.com/sunibgram/domain/model/thread"
	eventrepoiface "github.com/sunibgram/domain/repository/event"
	eventparticipantrepoiface "github.com/sunibgram/domain/repository/event_participant"
	eventparticipantsvciface "github.com/sunibgram/domain/service/event_participant"
)

type EventParticipantService struct {
	partRepo  eventparticipantrepoiface.IEventParticipantRepository
	eventRepo eventrepoiface.IEventRepository
}

func NewEventParticipantService(partRepo eventparticipantrepoiface.IEventParticipantRepository, eventRepo eventrepoiface.IEventRepository) eventparticipantsvciface.IEventParticipantService {
	return &EventParticipantService{partRepo: partRepo, eventRepo: eventRepo}
}

func (s *EventParticipantService) Join(userID uint, eventID string) error {
	if _, err := uuid.Parse(eventID); err != nil {
		return errors.New("invalid event id")
	}
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status != "approved" {
		return errors.New("event is not open for participation")
	}
	if !event.EndTime.IsZero() && event.EndTime.Before(time.Now().UTC()) {
		return errors.New("event has already ended")
	}
	exists, err := s.partRepo.Exists(userID, eventID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already joined")
	}
	return s.partRepo.Join(&eventparticipantmodel.EventParticipant{UserID: userID, EventID: eventID})
}

func (s *EventParticipantService) Leave(userID uint, eventID string) error {
	if _, err := uuid.Parse(eventID); err != nil {
		return errors.New("invalid event id")
	}
	exists, err := s.partRepo.Exists(userID, eventID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not joined")
	}
	return s.partRepo.Leave(userID, eventID)
}

func (s *EventParticipantService) GetMyJoined(userID uint, page, limit int) ([]threadmodel.Event, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.partRepo.FindMyJoined(userID, page, limit)
}
