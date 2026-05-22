package repository

import (
	"github.com/google/uuid"

	"github.com/sunibgram/domain/model"
)

type IViolationRepository interface {
	Save(violation *model.Violation) error
	FindAll() ([]model.Violation, error)
	FindByID(id uuid.UUID) (*model.Violation, error)
	Update(violation *model.Violation) error
	Delete(id uuid.UUID) error
}
