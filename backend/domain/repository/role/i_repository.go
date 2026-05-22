package rolerepoiface

import rolemodel "github.com/sunibgram/domain/model/role"

type IRoleRepository interface {
	Save(role *rolemodel.Role) error
	FindByID(id uint) (*rolemodel.Role, error)
	FindByName(name string) (*rolemodel.Role, error)
	FindAll() ([]rolemodel.Role, error)
	Update(role *rolemodel.Role) error
	Delete(id uint) error
}
