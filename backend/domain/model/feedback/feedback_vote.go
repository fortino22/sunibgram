package feedbackmodel

import (
	"time"

	"github.com/google/uuid"
)

type FeedbackVote struct {
	FeedbackID uuid.UUID `gorm:"type:uuid;primaryKey;autoIncrement:false"`
	UserID     uint      `gorm:"primaryKey;autoIncrement:false"`
	Value      int       `gorm:"type:smallint;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
