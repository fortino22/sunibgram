package usecase

import (
	"context"
	"errors"

	"gorm.io/gorm"

	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	usersettingssvciface "github.com/sunibgram/domain/service/user_settings"
)

type UserSettingsService struct {
	settingsRepo usersettingsrepoiface.IUserSettingsRepository
	threadRepo   threadrepoiface.IThreadRepository
}

func NewUserSettingsService(settingsRepo usersettingsrepoiface.IUserSettingsRepository, threadRepo threadrepoiface.IThreadRepository) usersettingssvciface.IUserSettingsService {
	return &UserSettingsService{settingsRepo: settingsRepo, threadRepo: threadRepo}
}

func (s *UserSettingsService) GetUserSettingByUserID(ctx context.Context, userID uint) (*usersettingsmodel.UserSettings, error) {
	setting, err := s.settingsRepo.GetUserSettingByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user settings not found")
	}
	return setting, err
}

func (s *UserSettingsService) UpdateUserSetting(ctx context.Context, userID uint, isPrivate bool) error {
	existing, err := s.settingsRepo.GetUserSettingByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil && existing.UserID != userID {
		return errors.New("only the owner can update this setting")
	}
	if err := s.settingsRepo.UpdateUserSetting(ctx, &usersettingsmodel.UserSettings{
		UserID:    userID,
		IsPrivate: isPrivate,
	}); err != nil {
		return err
	}
	if s.threadRepo != nil {
		s.threadRepo.InvalidateUserFeedCaches(ctx, userID)
	}
	return nil
}
