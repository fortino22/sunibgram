package categorysvciface

import (
	"github.com/google/uuid"

	categorymodel "github.com/sunibgram/domain/model/category"
)

type ICategoryService interface {
	Create(name string) error
	GetAll() ([]categorymodel.Category, error)
	Update(id uuid.UUID, name string) error
	Delete(id uuid.UUID) error
}
