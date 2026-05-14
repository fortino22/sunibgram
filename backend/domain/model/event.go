package model

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	ID                   string    `gorm:"type:varchar(36);primaryKey;column:id"`
	Title                string    `gorm:"type:varchar(255);not null"`
	Description          string    `gorm:"type:text"`
	Location             string    `gorm:"type:varchar(255)"`
	Type                 string    `gorm:"type:varchar(20);not null;default:'unofficial'"`
	Status               string    `gorm:"type:varchar(20);not null;default:'approved'"`
	StartTime            time.Time `gorm:"not null;index"`
	EndTime              time.Time `gorm:"not null;index"`
	BannerObjectName     string    `gorm:"type:varchar(500)"`
	BannerBucket         string    `gorm:"type:varchar(100)"`
	AttachmentObjectName string `gorm:"type:varchar(500)"`
	AttachmentBucket     string `gorm:"type:varchar(100)"`
	ParticipantCount     int       `gorm:"not null;default:0"`
	ThreadCount          int       `gorm:"not null;default:0"`
	LikeCount            int       `gorm:"not null;default:0"`
	RepostCount          int       `gorm:"not null;default:0"`
	WeekThreadCount      int       `gorm:"default:0"`
	WeekLikeCount        int       `gorm:"default:0"`
	WeekRepostCount      int       `gorm:"default:0"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

func ValidateEventType(t string) error {
	if t != "official" && t != "unofficial" {
		return errors.New("type must be official or unofficial")
	}
	return nil
}

var allowedEventAttachmentExts = map[string]bool{".zip": true}

func ValidateEventAttachmentExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedEventAttachmentExts[ext] {
		return errors.New("attachment must be a .zip file")
	}
	return nil
}

type ThreadEvent struct {
	ThreadID  uint      `gorm:"primaryKey;column:thread_id"`
	EventID   string    `gorm:"type:varchar(36);primaryKey;column:event_id"`
	CreatedAt time.Time `gorm:"type:timestamptz"`
	Thread    Thread    `gorm:"foreignKey:ThreadID;references:ThreadID;constraint:OnDelete:CASCADE"`
	Event     Event     `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE"`
}
