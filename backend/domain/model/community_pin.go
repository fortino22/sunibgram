package model

import "time"

type UserCommunityPin struct {
	UserID      uint `gorm:"primaryKey;autoIncrement:false;index:idx_user_community_pins_user,priority:1"`
	CommunityID uint `gorm:"primaryKey;autoIncrement:false;index:idx_user_community_pins_user,priority:2"`
	CreatedAt   time.Time
}
