package eventreviewmodel

import (
	"time"

	threadmodel "github.com/sunibgram/domain/model/thread"
	usermodel "github.com/sunibgram/domain/model/user"
)

type EventReview struct {
	ID        string            `gorm:"type:uuid;primaryKey;column:id"`
	EventID   string            `gorm:"type:varchar(36);not null;index"`
	AdminID   uint              `gorm:"not null;index"`
	Action    string            `gorm:"type:varchar(20);not null"`
	Note      string            `gorm:"type:text"`
	CreatedAt time.Time         `gorm:"type:timestamptz"`
	Event     threadmodel.Event `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE"`
	Admin     usermodel.User    `gorm:"foreignKey:AdminID;references:UserID"`
}
