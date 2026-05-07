package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IUserSettingsRepository interface {
	GetUserSettingByUserID(ctx context.Context, userID uint) (*model.UserSettings, error)
	IsUserPrivate(ctx context.Context, userID uint) (bool, error)
	UpdateUserSetting(ctx context.Context, setting *model.UserSettings) error
}
