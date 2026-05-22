package service

import (
	"github.com/google/uuid"

	"github.com/sunibgram/domain/model"
)

type IViolationService interface {
	Create(name, description string) error
	GetAll() ([]model.Violation, error)
	GetByID(id uuid.UUID) (*model.Violation, error)
	Update(id uuid.UUID, name, description string) error
	Delete(id uuid.UUID) error
}
