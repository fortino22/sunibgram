package notificationusermutesvciface

import (
	usermodel "github.com/sunibgram/domain/model/user"
)

type INotificationUserMuteService interface {
	Mute(userID, mutedUserID uint) error
	Unmute(userID, mutedUserID uint) error
	IsMuted(userID, mutedUserID uint) (bool, error)
	GetMutedUsers(userID uint, page, limit int) ([]usermodel.User, int64, error)
}
