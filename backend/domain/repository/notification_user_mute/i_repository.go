package notificationusermuterepoiface

type INotificationUserMuteRepository interface {
	Mute(userID, mutedUserID uint) error
	Unmute(userID, mutedUserID uint) error
	// IsMuted reports whether userID has muted mutedUserID (directional, one-way).
	IsMuted(userID, mutedUserID uint) (bool, error)
	// GetMutedIDs returns all user ids the given user mutes.
	GetMutedIDs(userID uint) ([]uint, error)
}
