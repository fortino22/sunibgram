package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IEventThreadPinService interface {
	Pin(ctx context.Context, eventID string, actorUserID, threadID uint) error
	Unpin(ctx context.Context, eventID string, actorUserID, threadID uint) error
	GetPinned(eventID string) (*model.EventThreadPin, error)
}
