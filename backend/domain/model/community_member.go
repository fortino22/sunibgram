package model

import "time"

type CommunityMemberRole string

const (
	CommunityMemberRoleMember    CommunityMemberRole = "MEMBER"
	CommunityMemberRoleModerator CommunityMemberRole = "MODERATOR"
	CommunityMemberRoleAdmin     CommunityMemberRole = "ADMIN"
	CommunityMemberRoleOwner     CommunityMemberRole = "OWNER"
)

type CommunityMember struct {
	CommunityID uint                `gorm:"primaryKey;autoIncrement:false"`
	UserID      uint                `gorm:"primaryKey;autoIncrement:false"`
	Community   Community           `gorm:"foreignKey:CommunityID;references:CommunityID"`
	User        User                `gorm:"foreignKey:UserID;references:UserID"`
	Role        CommunityMemberRole `gorm:"type:varchar(12);not null;default:'MEMBER'"`
	JoinedAt    time.Time
}
