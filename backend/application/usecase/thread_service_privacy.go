package usecase

import (
	"context"
	"errors"

	followrepoiface "github.com/sunibgram/domain/repository/follow"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
)

func canViewPrivateUser(
	ctx context.Context,
	settingsRepo usersettingsrepoiface.IUserSettingsRepository,
	followRepo followrepoiface.IFollowRepository,
	requesterID *uint,
	targetUserID uint,
) error {
	if requesterID != nil && *requesterID == targetUserID {
		return nil
	}
	isPrivate, err := settingsRepo.IsUserPrivate(ctx, targetUserID)
	if err != nil {
		return err
	}
	if !isPrivate {
		return nil
	}
	if requesterID == nil {
		return errors.New("this account is private")
	}
	following, err := followRepo.Exists(*requesterID, targetUserID)
	if err != nil {
		return err
	}
	if !following {
		return errors.New("this account is private")
	}
	return nil
}
