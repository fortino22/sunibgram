package eventactionconfirmationuc

import (
	"context"
	"time"

	eventactionconfirmationmodel "github.com/sunibgram/domain/model/event_action_confirmation"
	eventactionconfirmationrepoiface "github.com/sunibgram/domain/repository/event_action_confirmation"
	eventactionconfirmationsvciface "github.com/sunibgram/domain/service/event_action_confirmation"
)

type EventActionConfirmationService struct {
	repo eventactionconfirmationrepoiface.IEventActionConfirmationRepository
}

func NewEventActionConfirmationService(repo eventactionconfirmationrepoiface.IEventActionConfirmationRepository) eventactionconfirmationsvciface.IEventActionConfirmationService {
	return &EventActionConfirmationService{repo: repo}
}

func (s *EventActionConfirmationService) RecordEventDeletion(_ context.Context, eventID string, actorUserID uint, eventTitle string) error {
	rec := &eventactionconfirmationmodel.EventActionConfirmation{
		EventID:     eventID,
		ActorUserID: actorUserID,
		Type:        eventactionconfirmationmodel.ActionDeleteEvent,
		EventTitle:  eventTitle,
		CreatedAt:   time.Now(),
	}
	return s.repo.Save(rec)
}
