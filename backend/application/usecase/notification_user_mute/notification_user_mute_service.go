package notificationusermuteuc

import (
	"errors"

	"github.com/sunibgram/application/helper"
	usermodel "github.com/sunibgram/domain/model/user"
	notificationusermuterepoiface "github.com/sunibgram/domain/repository/notification_user_mute"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	notificationusermutesvciface "github.com/sunibgram/domain/service/notification_user_mute"
)

type NotificationUserMuteService struct {
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository
	userRepo userrepoiface.IUserRepository
}

func NewNotificationUserMuteService(
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository,
	userRepo userrepoiface.IUserRepository,
) notificationusermutesvciface.INotificationUserMuteService {
	return &NotificationUserMuteService{muteRepo: muteRepo, userRepo: userRepo}
}

func (s *NotificationUserMuteService) Mute(userID, mutedUserID uint) error {
	if userID == mutedUserID {
		return errors.New("cannot mute yourself")
	}
	if _, err := s.userRepo.FindByID(mutedUserID); err != nil {
		return errors.New("user not found")
	}
	// Mute is one-way and purely a visibility filter: no auto-unfollow, unlike block.
	return s.muteRepo.Mute(userID, mutedUserID)
}

func (s *NotificationUserMuteService) Unmute(userID, mutedUserID uint) error {
	return s.muteRepo.Unmute(userID, mutedUserID)
}

func (s *NotificationUserMuteService) IsMuted(userID, mutedUserID uint) (bool, error) {
	return s.muteRepo.IsMuted(userID, mutedUserID)
}

// GetMutedUsers returns the users that userID has muted, paginated.
// GetMutedIDs returns all muted IDs (small per-user set), so the page is sliced
// in memory; total is the full count for accurate pagination meta.
func (s *NotificationUserMuteService) GetMutedUsers(userID uint, page, limit int) ([]usermodel.User, int64, error) {
	ids, err := s.muteRepo.GetMutedIDs(userID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(ids))
	if total == 0 {
		return []usermodel.User{}, 0, nil
	}

	page, limit = helper.NormalizePageLimit(page, limit, 20)
	start := (page - 1) * limit
	if start >= len(ids) {
		return []usermodel.User{}, total, nil
	}
	end := start + limit
	if end > len(ids) {
		end = len(ids)
	}

	users, err := s.userRepo.FindByIDs(ids[start:end])
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
