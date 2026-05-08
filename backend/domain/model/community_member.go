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
	CommunityID uint                `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_community_member,priority:1"`
	UserID      uint                `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_community_member,priority:2"`
	Community   Community           `gorm:"foreignKey:CommunityID;references:CommunityID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	User        User                `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	Role        CommunityMemberRole `gorm:"type:varchar(12);not null;default:'MEMBER'"`
	JoinedAt    time.Time
}
