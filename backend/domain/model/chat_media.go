package model

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

var allowedChatMediaExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true,
}

var allowedChatMediaContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,

	"video/mp4":        true,
	"video/quicktime":  true,
	"video/x-msvideo":  true,
	"video/x-matroska": true,
	"video/webm":       true,
}

func ValidateChatMediaExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedChatMediaExts[ext] {
		return errors.New("only image and video files are allowed")
	}
	return nil
}

func ValidateChatMediaContentType(contentType string) error {
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if !allowedChatMediaContentTypes[mediaType] {
		return errors.New("only image and video files are allowed")
	}
	return nil
}

func ValidateChatMediaCount(count int) error {
	if count > 4 {
		return errors.New("maximum 4 media files allowed per message")
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
