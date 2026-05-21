package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/sunibgram/domain/model"
)

type IEventService interface {
	Create(ctx context.Context, userID uint, title, description, location string, eventType string, startTime, endTime time.Time, banner *multipart.FileHeader, attachments []*multipart.FileHeader) (*model.Event, error)
	GetByID(id string) (*model.Event, error)
	GetAll(limit int) ([]model.Event, []model.Event, []model.Event, error)
	GetPending(limit int) ([]model.Event, error)
	GetAdminEvents(status, name string, page, limit int) ([]model.Event, map[string][]model.EventReview, int64, error)
	GetMyEvents(userID uint, status string, page, limit int) ([]model.Event, map[string][]model.EventReview, int64, error)
	Update(ctx context.Context, id, title, description, location, eventType string, startTime, endTime time.Time, banner *multipart.FileHeader, clearBanner bool, attachments []*multipart.FileHeader) (*model.Event, error)
	ApproveEvent(id, adminID string) error
	DeclineEvent(id, adminID, reason string) error
	RequestRevision(id, adminID, note string) error
	Delete(ctx context.Context, id string) error
	GetThreads(ctx context.Context, requesterID *uint, id, filter, sort string, page, limit int) ([]model.Thread, int64, error)
	GetURL(bucket, objectName string) string
	GeneratePresignedAttachmentURL(ctx context.Context, bucket, objectName string) (string, error)
}
