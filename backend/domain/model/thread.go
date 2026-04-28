package model

import "time"

type Thread struct {
	ThreadID  uint    `gorm:"primaryKey;autoIncrement"`
	UserID    uint    `gorm:"not null;index"`
	User      User    `gorm:"foreignKey:UserID;references:UserID"`
	Content   string  `gorm:"type:text;not null"`
	ParentID  *uint   `gorm:"index"`
	RootID    *uint   `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
