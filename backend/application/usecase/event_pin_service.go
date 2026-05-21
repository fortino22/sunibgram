package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type EventPinService struct {
	pinRepo   repository.IEventPinRepository
	eventRepo repository.IEventRepository
}

func NewEventPinService(pinRepo repository.IEventPinRepository, eventRepo repository.IEventRepository) service.IEventPinService {
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
	return s.pinRepo.Pin(&model.UserEventPin{UserID: userID, EventID: eventID})
}

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

func (s *EventPinService) GetMyPins(userID uint, page, limit int) ([]model.Event, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.pinRepo.FindMyPins(userID, page, limit)
}
