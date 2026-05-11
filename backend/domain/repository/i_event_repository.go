package repository

import "github.com/sunibgram/domain/model"

type IEventRepository interface {
	Save(event *model.Event) error
	FindByID(id string) (*model.Event, error)
	FindAll(page, limit int) ([]model.Event, int64, error)
	Update(event *model.Event) error
	Delete(id string) error
	FindThreads(eventID string, page, limit int) ([]model.Thread, int64, error)
}
