package repostmodel

import "time"

type Repost struct {
	UserID    uint `gorm:"primaryKey;autoIncrement:false;index:idx_reposts_user_thread,priority:1"`
	ThreadID  uint `gorm:"primaryKey;autoIncrement:false;index:idx_reposts_user_thread,priority:2"`
	CreatedAt time.Time
}
