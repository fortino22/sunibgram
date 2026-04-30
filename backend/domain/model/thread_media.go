package model

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

var allowedThreadMediaExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".mp4": true, ".mov": true, ".avi": true,
}

func ValidateThreadMediaExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedThreadMediaExts[ext] {
		return errors.New("unsupported file type: " + filename)
	}
	return nil
}

type ThreadMedia struct {
	MediaID    uint   `gorm:"primaryKey;autoIncrement"`
	ThreadID   uint   `gorm:"not null;index"`
	Thread     Thread `gorm:"foreignKey:ThreadID;references:ThreadID"`
	ObjectName string `gorm:"not null"`
	Bucket     string `gorm:"not null"`
	CreatedAt  time.Time
}
