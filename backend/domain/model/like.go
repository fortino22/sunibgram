package model

import "time"

type Like struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false"`
	ThreadID  uint      `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
}
