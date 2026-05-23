package usersettingsrepoiface

import (
	"context"

	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
)

type IUserSettingsRepository interface {
	GetUserSettingByUserID(ctx context.Context, userID uint) (*usersettingsmodel.UserSettings, error)
	IsUserPrivate(ctx context.Context, userID uint) (bool, error)
	GetPrivacyMapByUserIDs(ctx context.Context, userIDs []uint) (map[uint]bool, error)
	UpdateUserSetting(ctx context.Context, setting *usersettingsmodel.UserSettings) error
	Create(ctx context.Context, setting *usersettingsmodel.UserSettings) error
}
