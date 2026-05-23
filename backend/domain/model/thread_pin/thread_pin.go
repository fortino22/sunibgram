package threadpinmodel

import "time"

type ThreadPin struct {
	UserID    uint `gorm:"primaryKey;autoIncrement:false"`
	ThreadID  uint `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
}
