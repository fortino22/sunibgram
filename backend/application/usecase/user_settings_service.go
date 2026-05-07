package usecase

import (
	"context"
	"errors"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"gorm.io/gorm"
)

type UserSettingsService struct {
	settingsRepo repository.IUserSettingsRepository
}

func NewUserSettingsService(settingsRepo repository.IUserSettingsRepository) service.IUserSettingsService {
	return &UserSettingsService{settingsRepo: settingsRepo}
}

func (s *UserSettingsService) GetUserSettingByUserID(ctx context.Context, userID uint) (*model.UserSettings, error) {
	setting, err := s.settingsRepo.GetUserSettingByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user settings not found")
	}
	return setting, err
}

func (s *UserSettingsService) UpdateUserSetting(ctx context.Context, userID uint, isPrivate bool) error {
	return s.settingsRepo.UpdateUserSetting(ctx, &model.UserSettings{
		UserID:    userID,
		IsPrivate: isPrivate,
	})
}
