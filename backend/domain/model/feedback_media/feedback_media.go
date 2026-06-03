package feedbackmediamodel

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MaxFeedbackMediaItems is the business limit for images attached to a feedback.
const MaxFeedbackMediaItems = 4

func ValidateFeedbackMediaCount(count int) error {
	if count > MaxFeedbackMediaItems {
		return fmt.Errorf("feedback can have at most %d media items", MaxFeedbackMediaItems)
	}
	return nil
}

type FeedbackMedia struct {
	MediaID    uint      `gorm:"primaryKey;autoIncrement"`
	FeedbackID uuid.UUID `gorm:"type:uuid;not null;index"`
	ObjectName string    `gorm:"not null"`
	CreatedAt  time.Time
}
