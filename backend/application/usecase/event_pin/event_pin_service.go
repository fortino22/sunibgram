package eventpinuc

import (
	"errors"

	"github.com/google/uuid"

	"github.com/sunibgram/application/helper"
	eventpinmodel "github.com/sunibgram/domain/model/event_pin"
	threadmodel "github.com/sunibgram/domain/model/thread"
	eventrepoiface "github.com/sunibgram/domain/repository/event"
	eventpinrepoiface "github.com/sunibgram/domain/repository/event_pin"
	eventpinsvciface "github.com/sunibgram/domain/service/event_pin"
)

type EventPinService struct {
	pinRepo   eventpinrepoiface.IEventPinRepository
	eventRepo eventrepoiface.IEventRepository
}

func NewEventPinService(pinRepo eventpinrepoiface.IEventPinRepository, eventRepo eventrepoiface.IEventRepository) eventpinsvciface.IEventPinService {
	return &EventPinService{pinRepo: pinRepo, eventRepo: eventRepo}
}

func (s *EventPinService) Pin(userID uint, eventID string) error {
	if _, err := uuid.Parse(eventID); err != nil {
		return errors.New("invalid event id")
	}
	if _, err := s.eventRepo.FindByID(eventID); err != nil {
		return errors.New("event not found")
	}
	exists, err := s.pinRepo.Exists(userID, eventID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already pinned")
	}
	count, err := s.pinRepo.CountByUser(userID)
	if err != nil {
		return err
	}
	if count >= maxEventPinsPerUser {
		return errors.New("maximum of 3 pinned events reached")
	}
	return s.pinRepo.Pin(&eventpinmodel.UserEventPin{UserID: userID, EventID: eventID})
}

const maxEventPinsPerUser = 3

func (s *EventPinService) Unpin(userID uint, eventID string) error {
	if _, err := uuid.Parse(eventID); err != nil {
		return errors.New("invalid event id")
	}
	exists, err := s.pinRepo.Exists(userID, eventID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not pinned")
	}
	return s.pinRepo.Unpin(userID, eventID)
}

func (s *EventPinService) GetMyPins(userID uint, page, limit int) ([]threadmodel.Event, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.pinRepo.FindMyPins(userID, page, limit)
}
