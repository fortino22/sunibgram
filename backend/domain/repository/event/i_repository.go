package eventrepoiface

import (
	"context"

	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IEventRepository interface {
	Save(event *threadmodel.Event) error
	FindByID(id string) (*threadmodel.Event, error)
	FindGrouped(limit int) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error)
	// FindGroupedApproved groups approved events; when locationID/region is set it
	// keeps only events whose audience includes that campus/region (or "All Regions").
	FindGroupedApproved(limit int, locationID *uint, region string) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error)
	FindTopRunningOrFuture(limit int) ([]threadmodel.Event, error)
	// ReplaceAudience overwrites an event's target campuses (region audience).
	ReplaceAudience(eventID string, locationIDs []uint) error
	// FindAudienceByEventIDs returns each event's target campuses, keyed by event id.
	FindAudienceByEventIDs(eventIDs []string) (map[string][]facultylocationmodel.FacultyLocation, error)
	FindPending(limit int) ([]threadmodel.Event, error)
	FindAdmin(status, name string, page, limit int) ([]threadmodel.Event, int64, error)
	FindByUserID(userID uint, status, name string, page, limit int) ([]threadmodel.Event, int64, error)
	Update(event *threadmodel.Event) error
	UpdateStatus(id, status string) error
	Delete(id string) error
	FindThreads(eventID string, ownerUserID uint, requesterID *uint, filter, sort string, page, limit int) ([]threadmodel.Thread, int64, error)
	IsThreadInEvent(eventID string, threadID uint) (bool, error)
	PatchEventCounterByThreadID(ctx context.Context, threadID uint, column string, delta int) error
}
