package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserSettingsRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewUserSettingsRepository(writeDB *gorm.DB, readDB *gorm.DB) domainrepo.IUserSettingsRepository {
	return &UserSettingsRepository{writeDB: writeDB, readDB: readDB}
}

func (r *UserSettingsRepository) GetUserSettingByUserID(ctx context.Context, userID uint) (*model.UserSettings, error) {
	var setting model.UserSettings
	err := r.readDB.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error
	return &setting, err
}

func (r *UserSettingsRepository) IsUserPrivate(ctx context.Context, userID uint) (bool, error) {
	var setting model.UserSettings
	err := r.readDB.WithContext(ctx).
		Model(&model.UserSettings{}).
		Select("is_private").
		Where("user_id = ?", userID).
		Scan(&setting).Error
	return setting.IsPrivate, err
}

func (r *UserSettingsRepository) UpdateUserSetting(ctx context.Context, setting *model.UserSettings) error {
	return r.writeDB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_private"}),
	}).Create(setting).Error
}
