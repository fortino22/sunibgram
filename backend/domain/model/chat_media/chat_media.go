package chatmediamodel

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	chatmodel "github.com/sunibgram/domain/model/chat"
	mediamodel "github.com/sunibgram/domain/model/media"
)

var allowedChatMediaExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true,
	".ogg": true, ".mp3": true, ".m4a": true, ".wav": true,
}

var allowedChatMediaContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,

	"video/mp4":        true,
	"video/quicktime":  true,
	"video/avi":        true,
	"video/x-msvideo":  true,
	"video/x-matroska": true,
	"video/webm":       true,

	"application/ogg": true,
	"audio/ogg":       true,
	"audio/webm":      true,
	"audio/mpeg":      true,
	"audio/mp4":       true,
	"audio/wave":      true,
	"audio/wav":       true,
	"audio/x-wav":     true,
}

func ValidateChatMediaExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedChatMediaExts[ext] {
		return errors.New("only image, video, and audio files are allowed")
	}
	return nil
}

func ValidateChatMediaContentType(contentType string) error {
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if !allowedChatMediaContentTypes[mediaType] {
		return errors.New("only image, video, and audio files are allowed")
	}
	return nil
}

func ValidateChatMediaFile(file multipart.File) (string, error) {
	contentType, err := mediamodel.DetectMultipartContentType(file)
	if err != nil {
		return "", err
	}
	if !allowedChatMediaContentTypes[contentType] {
		return "", fmt.Errorf("media content type %s is not allowed", contentType)
	}
	return contentType, nil
}

func IsAudioMediaType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(contentType), "audio/")
}

func ValidateChatMediaCount(count int) error {
	if count > 4 {
		return errors.New("maximum 4 media files allowed per message")
	}
	return nil
}

func (ChatMedia) TableName() string { return "chat_media" }

type ChatMedia struct {
	MediaID    uint           `gorm:"primaryKey;autoIncrement"`
	ChatID     uint           `gorm:"not null;index"`
	Chat       chatmodel.Chat `gorm:"foreignKey:ChatID;references:ChatID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	ObjectName string         `gorm:"not null"`
	Bucket     string         `gorm:"not null"`
	MediaType  string         `gorm:"type:varchar(50);not null"`
	FileSize   int64
	CreatedAt  time.Time
}
