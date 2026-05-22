package rolesvciface

import rolemodel "github.com/sunibgram/domain/model/role"

type IRoleService interface {
	Create(role *rolemodel.Role) error
	GetAll() ([]rolemodel.Role, error)
	GetByID(id uint) (*rolemodel.Role, error)
	Update(role *rolemodel.Role) error
	Delete(id uint) error
}
