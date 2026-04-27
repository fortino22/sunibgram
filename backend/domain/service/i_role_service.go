package service

import "github.com/sunibgram/domain/model"

type IRoleService interface {
	Create(role *model.Role) error
	GetAll() ([]model.Role, error)
	GetByID(id uint) (*model.Role, error)
	Update(role *model.Role) error
	Delete(id uint) error
}
