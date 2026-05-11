package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type EventService struct {
	eventRepo repository.IEventRepository
}

func NewEventService(eventRepo repository.IEventRepository) service.IEventService {
	return &EventService{eventRepo: eventRepo}
}

func (s *EventService) Create(ctx context.Context, title, description, location string, startTime, endTime time.Time) (*model.Event, error) {
	if err := validateEventInput(title, startTime, endTime); err != nil {
		return nil, err
	}
	event := &model.Event{
		Title:       strings.TrimSpace(title),
		Description: description,
		Location:    location,
		StartTime:   startTime.UTC(),
		EndTime:     endTime.UTC(),
	}
	if err := s.eventRepo.Save(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) GetByID(id string) (*model.Event, error) {
	if err := validateEventID(id); err != nil {
		return nil, err
	}
	return s.eventRepo.FindByID(id)
}

func (s *EventService) GetAll(page, limit int) ([]model.Event, int64, error) {
	page, limit = normalizePageLimit(page, limit)
	return s.eventRepo.FindAll(page, limit)
}

func (s *EventService) Update(ctx context.Context, id, title, description, location string, startTime, endTime time.Time) (*model.Event, error) {
	if err := validateEventID(id); err != nil {
		return nil, err
	}
	if err := validateEventInput(title, startTime, endTime); err != nil {
		return nil, err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("event not found")
	}
	event.Title = strings.TrimSpace(title)
	event.Description = description
	event.Location = location
	event.StartTime = startTime.UTC()
	event.EndTime = endTime.UTC()
	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	if err := validateEventID(id); err != nil {
		return err
	}
	if _, err := s.eventRepo.FindByID(id); err != nil {
		return errors.New("event not found")
	}
	return s.eventRepo.Delete(id)
}

func (s *EventService) GetThreads(ctx context.Context, id string, page, limit int) ([]model.Thread, int64, error) {
	if err := validateEventID(id); err != nil {
		return nil, 0, err
	}
	if _, err := s.eventRepo.FindByID(id); err != nil {
		return nil, 0, errors.New("event not found")
	}
	page, limit = normalizePageLimit(page, limit)
	return s.eventRepo.FindThreads(id, page, limit)
}

func validateEventID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid event id")
	}
	return nil
}

func validateEventInput(title string, startTime, endTime time.Time) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}
	if startTime.IsZero() {
		return errors.New("start_time is required")
	}
	if endTime.IsZero() {
		return errors.New("end_time is required")
	}
	if !endTime.After(startTime) {
		return errors.New("end_time must be after start_time")
	}
	return nil
}

func normalizePageLimit(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}
