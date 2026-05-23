package violationrepoiface

import (
	"github.com/google/uuid"

	violationmodel "github.com/sunibgram/domain/model/violation"
)

type IViolationRepository interface {
	Save(violation *violationmodel.Violation) error
	FindAll() ([]violationmodel.Violation, error)
	FindByID(id uuid.UUID) (*violationmodel.Violation, error)
	Update(violation *violationmodel.Violation) error
	Delete(id uuid.UUID) error
}
