package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/sunibgram/domain/model"
)

type IEventService interface {
	Create(ctx context.Context, title, description, location string, startTime, endTime time.Time, banner *multipart.FileHeader) (*model.Event, error)
	GetByID(id string) (*model.Event, error)
	GetAll(limit int) ([]model.Event, []model.Event, []model.Event, error)
	Update(ctx context.Context, id, title, description, location string, startTime, endTime time.Time, banner *multipart.FileHeader, clearBanner bool) (*model.Event, error)
	Delete(ctx context.Context, id string) error
	GetThreads(ctx context.Context, id string, page, limit int) ([]model.Thread, int64, error)
	GetURL(bucket, objectName string) string
}
