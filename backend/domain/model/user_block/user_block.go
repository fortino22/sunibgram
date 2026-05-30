package userblockmodel

import "time"

type UserBlock struct {
	BlockerID uint `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_user_block,priority:1"`
	BlockedID uint `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_user_block,priority:2;index:idx_user_block_blocked"`
	CreatedAt time.Time
}
