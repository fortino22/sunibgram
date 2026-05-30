package notificationusermutemodel

import "time"

// NotificationUserMute is a one-way mute: UserID has muted MutedUserID, so the
// muted user's posts and notifications are hidden from UserID only.
type NotificationUserMute struct {
	UserID      uint `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_notification_user_mute,priority:1"`
	MutedUserID uint `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_notification_user_mute,priority:2;index:idx_notification_user_mute_muted"`
	CreatedAt   time.Time
}
