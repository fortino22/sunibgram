package model

import "time"

type UserProfile struct {
	ProfileID          uint      `gorm:"primaryKey;autoIncrement"`
	UserID             uint      `gorm:"not null;uniqueIndex"`
	User               User      `gorm:"foreignKey:UserID;references:UserID"`
	AvatarObjectName   string    `gorm:"default:''"`
	AvatarBucket       string    `gorm:"default:''"`
	BannerObjectName   string    `gorm:"default:''"`
	BannerBucket       string    `gorm:"default:''"`
	Bio                string    `gorm:"type:text;default:''"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
