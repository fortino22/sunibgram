package model

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

// MaxThreadMediaItems is the business limit for media attached to a thread.
const MaxThreadMediaItems = 4

var allowedThreadMediaExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
	".mp4": true, ".mov": true, ".avi": true,
}

var allowedThreadMediaContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"image/gif":       true,
	"video/avi":       true,
	"video/mp4":       true,
	"video/quicktime": true,
	"video/x-msvideo": true,
}

func ValidateThreadMediaExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedThreadMediaExts[ext] {
		return errors.New("unsupported file type: " + filename)
	}
	return nil
}

func ValidateThreadMediaFile(file multipart.File) (string, error) {
	contentType, err := DetectMultipartContentType(file)
	if err != nil {
		return "", err
	}
	if !allowedThreadMediaContentTypes[contentType] {
		return "", fmt.Errorf("media content type %s is not allowed", contentType)
	}
	return contentType, nil
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
