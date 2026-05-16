package model

import (
	"time"

	"github.com/google/uuid"
)

type CommunityVisibility string

const (
	CommunityVisibilityPublic  CommunityVisibility = "PUBLIC"
	CommunityVisibilityPrivate CommunityVisibility = "PRIVATE"
)

type Community struct {
	CommunityID      uint                `gorm:"primaryKey;autoIncrement"`
	UserID           uint                `gorm:"not null;index;index:idx_community_creator,priority:1"`
	User             User                `gorm:"foreignKey:UserID;references:UserID"`
	Name             string              `gorm:"type:varchar(100);not null"`
	Description      string              `gorm:"type:text"`
	Rules            string              `gorm:"type:text"`
	CategoryID       *uuid.UUID          `gorm:"type:uuid;index"`
	Category         *Category           `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Visibility       CommunityVisibility `gorm:"type:varchar(10);not null;default:'PUBLIC';index"`
	AvatarObjectName string              `gorm:"type:varchar(500)"`
	AvatarBucket     string              `gorm:"type:varchar(100)"`
	BannerObjectName string              `gorm:"type:varchar(500)"`
	BannerBucket     string              `gorm:"type:varchar(100)"`
	MemberCount      int                 `gorm:"not null;default:1"`
	IsActive         bool                `gorm:"not null;default:true;index"`
	CreatedAt        time.Time           `gorm:"index:idx_community_creator,priority:2,sort:desc"`
	UpdatedAt        time.Time
}

func (c *Community) IsWritable() bool {
	return c.IsActive
}
