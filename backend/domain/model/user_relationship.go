package model

import "time"

type UserRelationship struct {
	FollowerID  uint      `gorm:"primaryKey;autoIncrement:false"`
	FollowingID uint      `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt   time.Time
}
