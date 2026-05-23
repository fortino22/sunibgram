package categoryrepoiface

import (
	"github.com/google/uuid"

	categorymodel "github.com/sunibgram/domain/model/category"
)

type ICategoryRepository interface {
	Save(category *categorymodel.Category) error
	FindAll() ([]categorymodel.Category, error)
	FindByID(id uuid.UUID) (*categorymodel.Category, error)
	Update(category *categorymodel.Category) error
	Delete(id uuid.UUID) error
}
