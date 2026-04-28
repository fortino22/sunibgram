package model

import "time"

type Bookmark struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false"`
	ThreadID  uint      `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
}
