package service

import (
	"context"

	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
)

type IUserSettingsService interface {
	GetUserSettingByUserID(ctx context.Context, userID uint) (*usersettingsmodel.UserSettings, error)
	UpdateUserSetting(ctx context.Context, userID uint, isPrivate bool) error
}
