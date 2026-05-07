package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IUserSettingsRepository interface {
	GetUserSettingByUserID(ctx context.Context, userID uint) (*model.UserSettings, error)
	UpdateUserSetting(ctx context.Context, setting *model.UserSettings) error
}
