package userrelationshipmodel

import "time"

type UserRelationship struct {
	FollowerID  uint `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_user_relationship,priority:1"`
	FollowingID uint `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_user_relationship,priority:2;index:idx_user_relationship_following"`
	CreatedAt   time.Time
}
