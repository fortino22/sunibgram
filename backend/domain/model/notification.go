package model

import "time"

type Notification struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	Type            string    `gorm:"type:varchar(50);not null"`
	Message         string    `gorm:"type:text;not null"`
	SenderUserID    uint      `gorm:"not null;index"`
	RecipientUserID uint      `gorm:"not null;index"`
	ReferenceID     *uint     `gorm:"index"`
	CreatedAt       time.Time `gorm:"index"`
}

type NotificationSender struct {
	UserID    uint     `json:"user_id"`
	Username  string   `json:"username"`
	Name      string   `json:"name"`
	AvatarURL string   `json:"avatar_url"`
	IsPremium bool     `json:"is_premium"`
	Faculty   *Faculty `json:"faculty,omitempty"`
}
