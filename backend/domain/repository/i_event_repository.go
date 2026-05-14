package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IEventRepository interface {
	Save(event *model.Event) error
	FindByID(id string) (*model.Event, error)
	FindGrouped(limit int) ([]model.Event, []model.Event, []model.Event, error)
	FindGroupedApproved(limit int) ([]model.Event, []model.Event, []model.Event, error)
	FindPending(limit int) ([]model.Event, error)
	Update(event *model.Event) error
	UpdateStatus(id, status string) error
	Delete(id string) error
	FindThreads(eventID string, page, limit int) ([]model.Thread, int64, error)
	PatchEventCounterByThreadID(ctx context.Context, threadID uint, column string, delta int) error
}
