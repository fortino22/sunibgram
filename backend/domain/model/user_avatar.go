package model

import "time"

type UserAvatar struct {
	AvatarID   uint      `gorm:"primaryKey;autoIncrement"`
	UserID     uint      `gorm:"not null;uniqueIndex"`
	User       User      `gorm:"foreignKey:UserID;references:UserID"`
	ObjectName string    `gorm:"not null"`
	Bucket     string    `gorm:"not null"`
	URL        string    `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
