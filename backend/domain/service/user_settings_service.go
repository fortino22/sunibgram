package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IUserSettingsService interface {
	GetUserSettingByUserID(ctx context.Context, userID uint) (*model.UserSettings, error)
	UpdateUserSetting(ctx context.Context, userID uint, isPrivate bool) error
}
