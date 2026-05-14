package model

import (
	"database/sql/driver"
	"encoding/json"
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
	Status               string    `gorm:"type:varchar(20);not null;default:'pending'"`
	StartTime            time.Time `gorm:"not null;index"`
	EndTime              time.Time `gorm:"not null;index"`
	BannerObjectName     string    `gorm:"type:varchar(500)"`
	BannerBucket         string    `gorm:"type:varchar(100)"`
	AttachmentObjectName string    `gorm:"type:varchar(500)"`
	AttachmentBucket     string    `gorm:"type:varchar(100)"`
	PDFs                 EventPDFs `gorm:"type:jsonb;default:'[]'"`
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

var allowedEventAttachmentExts = map[string]bool{".pdf": true}

type EventPDF struct {
	ObjectName string `json:"object_name"`
	Bucket     string `json:"bucket"`
}

type EventPDFs []EventPDF

func (p EventPDFs) Value() (driver.Value, error) {
	if len(p) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (p *EventPDFs) Scan(value any) error {
	if value == nil {
		*p = EventPDFs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return errors.New("invalid event pdfs value")
	}
}

func ValidateEventPDFCount(count int) error {
	if count > 4 {
		return errors.New("maximum 4 PDF files allowed per event")
	}
	return nil
}

func ValidateEventAttachmentExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedEventAttachmentExts[ext] {
		return errors.New("attachment must be a .pdf file")
	}
	return nil
}

func ValidateEventAttachmentContentType(contentType string) error {
	if strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) != "application/pdf" {
		return errors.New("attachment must be an application/pdf file")
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
