package communityeventmodel

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	userprofilemodel "github.com/sunibgram/domain/model/user_profile"

	usermodel "github.com/sunibgram/domain/model/user"
)

// CommunityEventComment is a comment on a community event. Content may embed
// @[username|id] mention tokens; mentioned users are notified.
type CommunityEventComment struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	CommunityEventID uint   `gorm:"not null;index"`
	UserID           uint   `gorm:"not null;index"`
	Content          string `gorm:"type:text;not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time

	User    usermodel.User                `gorm:"foreignKey:UserID;references:UserID"`
	Profile *userprofilemodel.UserProfile `gorm:"foreignKey:UserID;references:UserID;<-:false"`
}

func ValidateCommentContent(content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("comment is required")
	}
	if utf8.RuneCountInString(content) > 2000 {
		return errors.New("comment must be 2000 characters or less")
	}
	return nil
}
