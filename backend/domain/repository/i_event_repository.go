package repository

import "github.com/sunibgram/domain/model"

type IEventRepository interface {
	Save(event *model.Event) error
	FindByID(id string) (*model.Event, error)
	FindGrouped(limit int) ([]model.Event, []model.Event, []model.Event, error)
	Update(event *model.Event) error
	Delete(id string) error
	FindThreads(eventID string, page, limit int) ([]model.Thread, int64, error)
	IncrementViewCount(id string) error
}
