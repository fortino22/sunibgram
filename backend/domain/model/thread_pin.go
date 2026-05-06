package model

import "time"

type ThreadPin struct {
	UserID    uint   `gorm:"primaryKey;autoIncrement:false"`
	ThreadID  uint   `gorm:"primaryKey;autoIncrement:false"`
	Thread    Thread `gorm:"foreignKey:ThreadID;references:ThreadID"`
	CreatedAt time.Time
}
