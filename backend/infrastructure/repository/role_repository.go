package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"

	"gorm.io/gorm"
)

type RoleRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewRoleRepository(writeDB *gorm.DB, readDB *gorm.DB) domainrepo.IRoleRepository {
	return &RoleRepository{writeDB: writeDB, readDB: readDB}
}

func (r *RoleRepository) Save(role *model.Role) error {
	return r.writeDB.Create(role).Error
}

func (r *RoleRepository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.readDB.First(&role, id).Error
	return &role, err
}

func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.readDB.Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *RoleRepository) FindAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.readDB.Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) Update(role *model.Role) error {
	return r.writeDB.Save(role).Error
}

func (r *RoleRepository) Delete(id uint) error {
	return r.writeDB.Delete(&model.Role{}, id).Error
}
