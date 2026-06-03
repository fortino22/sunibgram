package eventsvciface

import (
	"context"
	"mime/multipart"
	"time"

	eventreviewmodel "github.com/sunibgram/domain/model/event_review"
	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

type IEventService interface {
	Create(ctx context.Context, userID uint, title, description, location string, eventType string, startTime, endTime time.Time, banner *multipart.FileHeader, attachments []*multipart.FileHeader, facultyLocationIDs []uint) (*threadmodel.Event, error)
	GetByID(id string) (*threadmodel.Event, error)
	GetByIDWithReviews(id string) (*threadmodel.Event, []eventreviewmodel.EventReview, error)
	GetAll(limit int, locationID *uint, region string) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error)
	GetTopRunningOrFuture(limit int) ([]threadmodel.Event, error)
	// GetAudienceByEventIDs returns each event's target campuses, keyed by event id.
	GetAudienceByEventIDs(eventIDs []string) (map[string][]facultylocationmodel.FacultyLocation, error)
	GetPending(limit int) ([]threadmodel.Event, error)
	GetAdminEvents(status, name string, page, limit int) ([]threadmodel.Event, map[string][]eventreviewmodel.EventReview, int64, error)
	GetMyEvents(userID uint, status, name string, page, limit int) ([]threadmodel.Event, map[string][]eventreviewmodel.EventReview, int64, error)
	Update(ctx context.Context, id, title, description, location, eventType string, startTime, endTime time.Time, banner *multipart.FileHeader, clearBanner bool, attachments []*multipart.FileHeader, facultyLocationIDs []uint) (*threadmodel.Event, error)
	ApproveEvent(id, adminID string) error
	DeclineEvent(id, adminID, reason string) error
	RequestRevision(id, adminID, note string) error
	Delete(ctx context.Context, id string) error
	GetThreads(ctx context.Context, requesterID *uint, id, filter, sort string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error)
	GetURL(objectName string) string
	GeneratePresignedAttachmentURL(ctx context.Context, objectName string) (string, error)
}
