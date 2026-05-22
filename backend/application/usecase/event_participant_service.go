package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/sunibgram/application/helper"
	eventparticipantmodel "github.com/sunibgram/domain/model/event_participant"
	threadmodel "github.com/sunibgram/domain/model/thread"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type EventParticipantService struct {
	partRepo  repository.IEventParticipantRepository
	eventRepo repository.IEventRepository
}

func NewEventParticipantService(partRepo repository.IEventParticipantRepository, eventRepo repository.IEventRepository) service.IEventParticipantService {
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
