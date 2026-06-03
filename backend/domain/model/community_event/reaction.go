package communityeventmodel

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// CommunityEventReaction is a single emoji reaction by a user on an event. The
// composite PK (CommunityEventID, UserID) enforces ONE reaction per user — a
// re-pick UPDATEs the emoji rather than inserting a second row.
type CommunityEventReaction struct {
	CommunityEventID uint   `gorm:"primaryKey;autoIncrement:false;index"`
	UserID           uint   `gorm:"primaryKey;autoIncrement:false"`
	Emoji            string `gorm:"type:varchar(32);not null"`
	CreatedAt        time.Time
}

func ValidateEmoji(emoji string) error {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return errors.New("emoji is required")
	}
	if utf8.RuneCountInString(emoji) > 16 {
		return errors.New("emoji is invalid")
	}
	return nil
}
