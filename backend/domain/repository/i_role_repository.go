package repository

import "github.com/sunibgram/domain/model"

type IRoleRepository interface {
	Save(role *model.Role) error
	FindByID(id uint) (*model.Role, error)
	FindByName(name string) (*model.Role, error)
	FindAll() ([]model.Role, error)
	Update(role *model.Role) error
	Delete(id uint) error
}
