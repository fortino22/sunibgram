package communityeventmodel

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	communitymodel "github.com/sunibgram/domain/model/community"
	usermodel "github.com/sunibgram/domain/model/user"
)

// CommunityEvent is an event scoped to a single community. Only a community
// ADMIN can create one. Visibility is limited to its tagged invitees (plus
// community admins). Unlike the global Event it has no document attachment and
// no approval workflow — it is published immediately. Uses a uint PK so it
// integrates with Notification.ReferenceID and the uint-keyed community world.
type CommunityEvent struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	CommunityID      uint      `gorm:"not null;index"`
	CreatedByUserID  uint      `gorm:"not null;index"`
	Title            string    `gorm:"type:varchar(255);not null"`
	Description      string    `gorm:"type:text"`
	Location         string    `gorm:"type:varchar(255)"`
	StartTime        time.Time `gorm:"not null;index"`
	EndTime          time.Time `gorm:"not null;index"`
	BannerObjectName string    `gorm:"type:varchar(500)"`
	// PinnedCommentID is the single comment an admin has pinned (nil = none).
	PinnedCommentID *uint `gorm:"index"`
	CommentCount    int   `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Community communitymodel.Community `gorm:"foreignKey:CommunityID;references:CommunityID"`
	Creator   usermodel.User           `gorm:"foreignKey:CreatedByUserID;references:UserID"`
}

func ValidateCommunityEventInput(title string, startTime, endTime time.Time) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(title) > 255 {
		return errors.New("title must be 255 characters or less")
	}
	if startTime.IsZero() {
		return errors.New("start_time is required")
	}
	if endTime.IsZero() {
		return errors.New("end_time is required")
	}
	if !endTime.After(startTime) {
		return errors.New("end_time must be after start_time")
	}
	return nil
}
