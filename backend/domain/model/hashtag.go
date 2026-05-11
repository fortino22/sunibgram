package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Hashtag struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ThreadCount int64     `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"type:timestamptz"`
}

func (h *Hashtag) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

type ThreadHashtag struct {
	ThreadID  uint      `gorm:"primaryKey;column:thread_id"`
	HashtagID uuid.UUID `gorm:"type:uuid;primaryKey;column:hashtag_id"`
	CreatedAt time.Time `gorm:"type:timestamptz"`
	Thread    Thread    `gorm:"foreignKey:ThreadID;references:ThreadID;constraint:OnDelete:CASCADE"`
	Hashtag   Hashtag   `gorm:"foreignKey:HashtagID;references:ID;constraint:OnDelete:CASCADE"`
}

type TrendingHashtag struct {
	Name        string  `json:"name"`
	ThreadCount int64   `json:"thread_count"`
	Score       float64 `json:"score"`
}
