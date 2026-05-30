package threadmodel

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
)

type Event struct {
	ID                   string    `gorm:"type:varchar(36);primaryKey;column:id"`
	UserID               uint      `gorm:"not null;index"`
	Title                string    `gorm:"type:varchar(255);not null"`
	Description          string    `gorm:"type:text"`
	Location             string    `gorm:"type:varchar(255)"`
	Type                 string    `gorm:"type:varchar(20);not null;default:'unofficial'"`
	Status               string    `gorm:"type:varchar(20);not null;default:'pending'"`
	StartTime            time.Time `gorm:"not null;index"`
	EndTime              time.Time `gorm:"not null;index"`
	BannerObjectName     string    `gorm:"type:varchar(500)"`
	AttachmentObjectName string    `gorm:"type:varchar(500)"`
	Docs                 EventDocs `gorm:"type:jsonb;default:'[]';column:docs"`
	ParticipantCount     int       `gorm:"not null;default:0"`
	ThreadCount          int       `gorm:"not null;default:0"`
	LikeCount            int       `gorm:"not null;default:0"`
	RepostCount          int       `gorm:"not null;default:0"`
	ShareCount           int       `gorm:"not null;default:0"`
	WeekThreadCount      int       `gorm:"default:0"`
	WeekLikeCount        int       `gorm:"default:0"`
	WeekRepostCount      int       `gorm:"default:0"`
	WeekShareCount       int       `gorm:"default:0"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	User                 usermodel.User `gorm:"foreignKey:UserID;references:UserID"`
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

var allowedEventDocExts = map[string]bool{".doc": true, ".docx": true}

type EventDoc struct {
	ObjectName string `json:"object_name"`
}

type EventDocs []EventDoc

func (p EventDocs) Value() (driver.Value, error) {
	if len(p) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (p *EventDocs) Scan(value any) error {
	if value == nil {
		*p = EventDocs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return errors.New("invalid event docs value")
	}
}

func ValidateEventDocumentRequired(count int) error {
	if count == 0 {
		return errors.New("document is required")
	}
	if count > 1 {
		return errors.New("only one document is allowed per event")
	}
	return nil
}

func ValidateEventDocExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedEventDocExts[ext] {
		return errors.New("document must be a .doc or .docx file")
	}
	return nil
}

func ValidateEventDocFile(_ multipart.File) error {
	return nil
}

type ThreadEvent struct {
	ThreadID   uint      `gorm:"primaryKey;column:thread_id"`
	EventID    string    `gorm:"type:varchar(36);primaryKey;column:event_id"`
	IsOfficial bool      `gorm:"not null;default:false;column:is_official"`
	CreatedAt  time.Time `gorm:"type:timestamptz"`
	Thread     Thread    `gorm:"foreignKey:ThreadID;references:ThreadID;constraint:OnDelete:CASCADE"`
	Event      Event     `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE"`
}
