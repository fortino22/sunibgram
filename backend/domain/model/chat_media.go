package model

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

var allowedChatMediaExts = map[string]bool{
	".pdf": true,
}

func ValidateChatMediaExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedChatMediaExts[ext] {
		return errors.New("only PDF files are allowed")
	}
	return nil
}

func ValidateChatMediaContentType(contentType string) error {
	if strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) != "application/pdf" {
		return errors.New("only PDF files are allowed")
	}
	return nil
}

func ValidateChatMediaCount(count int) error {
	if count > 4 {
		return errors.New("maximum 4 PDF files allowed per message")
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
