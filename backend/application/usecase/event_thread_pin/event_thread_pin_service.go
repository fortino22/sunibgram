package eventthreadpinuc

import (
	"context"
	"errors"

	eventthreadpinmodel "github.com/sunibgram/domain/model/event_thread_pin"
	eventrepoiface "github.com/sunibgram/domain/repository/event"
	eventthreadpinrepoiface "github.com/sunibgram/domain/repository/event_thread_pin"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	eventthreadpinsvciface "github.com/sunibgram/domain/service/event_thread_pin"
)

type EventThreadPinService struct {
	pinRepo    eventthreadpinrepoiface.IEventThreadPinRepository
	eventRepo  eventrepoiface.IEventRepository
	threadRepo threadrepoiface.IThreadRepository
}

func NewEventThreadPinService(
	pinRepo eventthreadpinrepoiface.IEventThreadPinRepository,
	eventRepo eventrepoiface.IEventRepository,
	threadRepo threadrepoiface.IThreadRepository,
) eventthreadpinsvciface.IEventThreadPinService {
	return &EventThreadPinService{pinRepo: pinRepo, eventRepo: eventRepo, threadRepo: threadRepo}
}

func (s *EventThreadPinService) Pin(ctx context.Context, eventID string, actorUserID, threadID uint) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status != "approved" {
		return errors.New("event is not approved")
	}
	if event.UserID != actorUserID {
		return errors.New("forbidden")
	}
	if _, err := s.threadRepo.FindByID(threadID); err != nil {
		return errors.New("thread not found")
	}
	inEvent, err := s.eventRepo.IsThreadInEvent(eventID, threadID)
	if err != nil {
		return err
	}
	if !inEvent {
		return errors.New("thread does not belong to this event")
	}

	existing, err := s.pinRepo.FindByEventID(eventID)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.ThreadID == threadID {
			return errors.New("already pinned")
		}
		if err := s.pinRepo.Unpin(eventID, existing.ThreadID); err != nil {
			return err
		}
	}
	return s.pinRepo.Pin(&eventthreadpinmodel.EventThreadPin{EventID: eventID, ThreadID: threadID})
}

func (s *EventThreadPinService) Unpin(ctx context.Context, eventID string, actorUserID, threadID uint) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status != "approved" {
		return errors.New("event is not approved")
	}
	if event.UserID != actorUserID {
		return errors.New("forbidden")
	}
	exists, err := s.pinRepo.Exists(eventID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not pinned")
	}
	return s.pinRepo.Unpin(eventID, threadID)
}

func (s *EventThreadPinService) GetPinned(eventID string) (*eventthreadpinmodel.EventThreadPin, error) {
	return s.pinRepo.FindByEventID(eventID)
}
