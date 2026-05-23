package communityrulemodel

import (
	"time"

	communitymodel "github.com/sunibgram/domain/model/community"
)

type CommunityRule struct {
	ID          uint                     `gorm:"primaryKey;autoIncrement"`
	CommunityID uint                     `gorm:"not null;index;uniqueIndex:idx_community_rule_order,priority:1"`
	Community   communitymodel.Community `gorm:"foreignKey:CommunityID;references:CommunityID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	OrderNum    int                      `gorm:"not null;uniqueIndex:idx_community_rule_order,priority:2"`
	Text        string                   `gorm:"type:varchar(500);not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
