package communityeventmodel

import (
	"time"

	userprofilemodel "github.com/sunibgram/domain/model/user_profile"

	usermodel "github.com/sunibgram/domain/model/user"
)

// CommunityEventInvitee is a tagged person on a community event. Invitees are
// the only non-admin users who can see and comment on the event.
type CommunityEventInvitee struct {
	CommunityEventID uint `gorm:"primaryKey;autoIncrement:false;index"`
	UserID           uint `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt        time.Time

	User    usermodel.User                `gorm:"foreignKey:UserID;references:UserID"`
	Profile *userprofilemodel.UserProfile `gorm:"foreignKey:UserID;references:UserID;<-:false"`
}
