package communitymodel

import (
	"time"

	"github.com/google/uuid"

	categorymodel "github.com/sunibgram/domain/model/category"
	communitydiscovermodel "github.com/sunibgram/domain/model/community_discover"
	usermodel "github.com/sunibgram/domain/model/user"
)

type CommunityVisibility string

const (
	CommunityVisibilityPublic  CommunityVisibility = "PUBLIC"
	CommunityVisibilityPrivate CommunityVisibility = "PRIVATE"
)

type Community struct {
	CommunityID      uint                                             `gorm:"primaryKey;autoIncrement"`
	UserID           uint                                             `gorm:"not null;index;index:idx_community_creator,priority:1"`
	User             usermodel.User                                   `gorm:"foreignKey:UserID;references:UserID"`
	Name             string                                           `gorm:"type:varchar(100);not null"`
	Description      string                                           `gorm:"type:text"`
	CategoryID       *uuid.UUID                                       `gorm:"type:uuid;index"`
	Category         *categorymodel.Category                          `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Visibility       CommunityVisibility                              `gorm:"type:varchar(10);not null;default:'PUBLIC';index"`
	BannerObjectName string                                           `gorm:"type:varchar(500)"`
	MemberCount      int                                              `gorm:"not null;default:1"`
	IsActive         bool                                             `gorm:"not null;default:true;index"`
	TopMembers       []communitydiscovermodel.DiscoverCommunityMember `gorm:"-"`
	CreatedAt        time.Time                                        `gorm:"index:idx_community_creator,priority:2,sort:desc"`
	UpdatedAt        time.Time
}

func (c *Community) IsWritable() bool {
	return c.IsActive
}
