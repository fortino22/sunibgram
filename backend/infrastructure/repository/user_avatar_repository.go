package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type UserAvatarRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewUserAvatarRepository(writeDB *gorm.DB, readDB *gorm.DB) domainrepo.IUserAvatarRepository {
	return &UserAvatarRepository{writeDB: writeDB, readDB: readDB}
}

func (r *UserAvatarRepository) Save(avatar *model.UserAvatar) error {
	return r.writeDB.Create(avatar).Error
}

func (r *UserAvatarRepository) Update(avatar *model.UserAvatar) error {
	return r.writeDB.Save(avatar).Error
}

func (r *UserAvatarRepository) FindByUserID(userID uint) (*model.UserAvatar, error) {
	var avatar model.UserAvatar
	err := r.readDB.Where("user_id = ?", userID).First(&avatar).Error
	return &avatar, err
}

func (r *UserAvatarRepository) ExistsByUserID(userID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.UserAvatar{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func (r *UserAvatarRepository) DeleteByUserID(userID uint) error {
	return r.writeDB.Where("user_id = ?", userID).Delete(&model.UserAvatar{}).Error
}
