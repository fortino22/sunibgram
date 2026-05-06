package model

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

var allowedChatMediaExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".mp4": true, ".mov": true,
}

func ValidateChatMediaExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedChatMediaExts[ext] {
		return errors.New("unsupported file type: " + filename)
	}
	return nil
}

func (ChatMedia) TableName() string { return "chat_media" }

type ChatMedia struct {
	MediaID    uint   `gorm:"primaryKey;autoIncrement"`
	ChatID     uint   `gorm:"not null;index"`
	Chat       Chat   `gorm:"foreignKey:ChatID;references:ChatID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	ObjectName string `gorm:"not null"`
	Bucket     string `gorm:"not null"`
	MediaType  string `gorm:"type:varchar(50);not null"`
	FileSize   int64
	CreatedAt  time.Time
}
