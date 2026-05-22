package violationsvciface

import (
	"github.com/google/uuid"

	violationmodel "github.com/sunibgram/domain/model/violation"
)

type IViolationService interface {
	Create(name, description string) error
	GetAll() ([]violationmodel.Violation, error)
	GetByID(id uuid.UUID) (*violationmodel.Violation, error)
	Update(id uuid.UUID, name, description string) error
	Delete(id uuid.UUID) error
}
