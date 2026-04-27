package usecase

import (
	"errors"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type RoleService struct {
	roleRepo repository.IRoleRepository
}

func NewRoleService(roleRepo repository.IRoleRepository) service.IRoleService {
	return &RoleService{roleRepo: roleRepo}
}

func (s *RoleService) Create(role *model.Role) error {
	if role.Name == "" {
		return errors.New("role name is required")
	}
	return s.roleRepo.Save(role)
}

func (s *RoleService) GetAll() ([]model.Role, error) {
	return s.roleRepo.FindAll()
}

func (s *RoleService) GetByID(id uint) (*model.Role, error) {
	return s.roleRepo.FindByID(id)
}

func (s *RoleService) Update(role *model.Role) error {
	if role.Name == "" {
		return errors.New("role name is required")
	}
	return s.roleRepo.Update(role)
}

func (s *RoleService) Delete(id uint) error {
	return s.roleRepo.Delete(id)
}
