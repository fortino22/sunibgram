package model

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// MaxThreadMediaItems is the business limit for media attached to a thread.
const MaxThreadMediaItems = 4

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

func ValidateThreadMediaCount(count int) error {
	if count > MaxThreadMediaItems {
		return fmt.Errorf("thread can have at most %d media items", MaxThreadMediaItems)
	}
	return nil
}

type ThreadMedia struct {
	MediaID    uint   `gorm:"primaryKey;autoIncrement"`
	ThreadID   uint   `gorm:"not null;index"`
	ObjectName string `gorm:"not null"`
	Bucket     string `gorm:"not null"`
	CreatedAt  time.Time
}

type CommunityThreadMedia struct {
	ThreadID uint
	Media    []ThreadMedia
}
