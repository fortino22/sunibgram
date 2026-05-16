package service

import (
	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
)

type ICategoryService interface {
	Create(name string) error
	GetAll() ([]model.Category, error)
	Update(id uuid.UUID, name string) error
	Delete(id uuid.UUID) error
}
