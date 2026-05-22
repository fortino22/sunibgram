package repository

import (
	"github.com/google/uuid"

	"github.com/sunibgram/domain/model"
)

type ICategoryRepository interface {
	Save(category *model.Category) error
	FindAll() ([]model.Category, error)
	FindByID(id uuid.UUID) (*model.Category, error)
	Update(category *model.Category) error
	Delete(id uuid.UUID) error
}
