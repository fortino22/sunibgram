package eventrepoiface

import (
	"context"

	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IEventRepository interface {
	Save(event *threadmodel.Event) error
	FindByID(id string) (*threadmodel.Event, error)
	FindGrouped(limit int) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error)
	FindGroupedApproved(limit int) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error)
	FindPending(limit int) ([]threadmodel.Event, error)
	FindAdmin(status, name string, page, limit int) ([]threadmodel.Event, int64, error)
	FindByUserID(userID uint, status string, page, limit int) ([]threadmodel.Event, int64, error)
	Update(event *threadmodel.Event) error
	UpdateStatus(id, status string) error
	Delete(id string) error
	FindThreads(eventID string, ownerUserID uint, requesterID *uint, filter, sort string, page, limit int) ([]threadmodel.Thread, int64, error)
	IsThreadInEvent(eventID string, threadID uint) (bool, error)
	PatchEventCounterByThreadID(ctx context.Context, threadID uint, column string, delta int) error
}
