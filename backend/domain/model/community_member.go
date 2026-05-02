package model

import "time"

type CommunityMemberRole string

const (
	CommunityMemberRoleMember CommunityMemberRole = "MEMBER"
	CommunityMemberRoleAdmin  CommunityMemberRole = "ADMIN"
	CommunityMemberRoleOwner  CommunityMemberRole = "OWNER"
)

type CommunityMember struct {
	CommunityID uint                `gorm:"primaryKey;autoIncrement:false"`
	UserID      uint                `gorm:"primaryKey;autoIncrement:false"`
	Community   Community           `gorm:"foreignKey:CommunityID;references:CommunityID"`
	User        User                `gorm:"foreignKey:UserID;references:UserID"`
	Role        CommunityMemberRole `gorm:"type:varchar(10);not null;default:'MEMBER'"`
	JoinedAt    time.Time
}
