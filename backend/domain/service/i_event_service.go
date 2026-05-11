package service

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
)

type IEventService interface {
	Create(ctx context.Context, title, description, location string, startTime, endTime time.Time) (*model.Event, error)
	GetByID(id string) (*model.Event, error)
	GetAll(page, limit int) ([]model.Event, int64, error)
	Update(ctx context.Context, id, title, description, location string, startTime, endTime time.Time) (*model.Event, error)
	Delete(ctx context.Context, id string) error
	GetThreads(ctx context.Context, id string, page, limit int) ([]model.Thread, int64, error)
}
