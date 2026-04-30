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
	err := r.readDB.Preload("Role").First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Save(user *model.User) error {
	return r.writeDB.Create(user).Error
}

func (r *UserRepository) SaveWithProfile(user *model.User, profile *model.UserProfile) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UserID = user.UserID
		return tx.Create(profile).Error
	})
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.readDB.Preload("Role").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByUserName(username string) (*model.User, error) {
	var user model.User
	err := r.readDB.Preload("Role").Where("user_name = ?", username).First(&user).Error
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

func (r *UserRepository) UpdatePassword(userID uint, hashedPassword string) error {
	return r.writeDB.Model(&model.User{}).Where("user_id = ?", userID).Update("password", hashedPassword).Error
}

func (r *UserRepository) UpdateInfo(userID uint, name, username string) error {
	return r.writeDB.Model(&model.User{}).Where("user_id = ?", userID).
		Updates(map[string]any{"name": name, "user_name": username}).Error
}

func (r *UserRepository) FindAllPaginated(page, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	offset := (page - 1) * limit

	if err := r.readDB.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.readDB.Preload("Role").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}
