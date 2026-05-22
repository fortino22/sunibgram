package communitymembermodel

import (
	"time"

	communitymodel "github.com/sunibgram/domain/model/community"

	usermodel "github.com/sunibgram/domain/model/user"

	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
)

type CommunityMemberRole string

const (
	CommunityMemberRoleMember CommunityMemberRole = "MEMBER"
	CommunityMemberRoleAdmin  CommunityMemberRole = "ADMIN"
)

type CommunityMember struct {
	CommunityID uint                          `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_community_member,priority:1"`
	UserID      uint                          `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_community_member,priority:2"`
	Community   communitymodel.Community      `gorm:"foreignKey:CommunityID;references:CommunityID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	User        usermodel.User                `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	Profile     *userprofilemodel.UserProfile `gorm:"foreignKey:UserID;references:UserID;<-:false"`
	Role        CommunityMemberRole           `gorm:"type:varchar(12);not null;default:'MEMBER'"`
	JoinedAt    time.Time
}
