package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	ID               string    `gorm:"type:varchar(36);primaryKey;column:id"`
	Title            string    `gorm:"type:varchar(255);not null"`
	Description      string    `gorm:"type:text"`
	Location         string    `gorm:"type:varchar(255)"`
	StartTime        time.Time `gorm:"not null;index"`
	EndTime          time.Time `gorm:"not null;index"`
	BannerObjectName string    `gorm:"type:varchar(500)"`
	BannerBucket     string    `gorm:"type:varchar(100)"`
	ParticipantCount int       `gorm:"not null;default:0"`
	ThreadCount      int       `gorm:"not null;default:0"`
	LikeCount        int       `gorm:"not null;default:0"`
	RepostCount      int       `gorm:"not null;default:0"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
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
