package repository

import (
	domainrepo "github.com/sunibgram/domain/repository"

	"gorm.io/gorm"

	rolemodel "github.com/sunibgram/domain/model/role"
)

type RoleRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewRoleRepository(writeDB *gorm.DB, readDB *gorm.DB) domainrepo.IRoleRepository {
	return &RoleRepository{writeDB: writeDB, readDB: readDB}
}

func (r *RoleRepository) Save(role *rolemodel.Role) error {
	return r.writeDB.Create(role).Error
}

func (r *RoleRepository) FindByID(id uint) (*rolemodel.Role, error) {
	var role rolemodel.Role
	err := r.readDB.First(&role, id).Error
	return &role, err
}

func (r *RoleRepository) FindByName(name string) (*rolemodel.Role, error) {
	var role rolemodel.Role
	err := r.readDB.Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *RoleRepository) FindAll() ([]rolemodel.Role, error) {
	var roles []rolemodel.Role
	err := r.readDB.Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) Update(role *rolemodel.Role) error {
	return r.writeDB.Save(role).Error
}

func (r *RoleRepository) Delete(id uint) error {
	return r.writeDB.Delete(&rolemodel.Role{}, id).Error
}
