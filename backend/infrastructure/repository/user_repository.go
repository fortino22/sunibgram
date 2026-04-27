package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"

	"gorm.io/gorm"
)

type UserRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewUserRepository(writeDB *gorm.DB, readDB *gorm.DB) domainrepo.IUserRepository {
	return &UserRepository{writeDB: writeDB, readDB: readDB}
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.readDB.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Save(user *model.User) error {
	return r.writeDB.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.readDB.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByUserName(username string) (*model.User, error) {
	var user model.User
	err := r.readDB.Where("user_name = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByUserName(username string) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.User{}).Where("user_name = ?", username).Count(&count).Error
	return count > 0, err
}
