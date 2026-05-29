package trendreportmodel

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	usermodel "github.com/sunibgram/domain/model/user"
)

type TrendReportReason struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code      string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	Label     string    `gorm:"type:varchar(120);not null"`
	CreatedAt time.Time `gorm:"type:timestamptz"`
	UpdatedAt time.Time `gorm:"type:timestamptz"`
}

func (r *TrendReportReason) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type TrendReport struct {
	ID         uuid.UUID            `gorm:"type:uuid;primaryKey"`
	ReporterID uint                 `gorm:"not null;index;uniqueIndex:uniq_trend_report_user_hashtag"`
	Reporter   usermodel.User       `gorm:"foreignKey:ReporterID;references:UserID;constraint:OnDelete:CASCADE"`
	HashtagID  uuid.UUID            `gorm:"type:uuid;not null;index;uniqueIndex:uniq_trend_report_user_hashtag"`
	Hashtag    hashtagmodel.Hashtag `gorm:"foreignKey:HashtagID;references:ID;constraint:OnDelete:CASCADE"`
	ReasonID   uuid.UUID            `gorm:"type:uuid;not null;index"`
	Reason     TrendReportReason    `gorm:"foreignKey:ReasonID;references:ID;constraint:OnDelete:RESTRICT"`
	CreatedAt  time.Time            `gorm:"type:timestamptz"`
}

func (r *TrendReport) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
