package communityrequestmodel

import (
	"time"

	communitymodel "github.com/sunibgram/domain/model/community"
	usermodel "github.com/sunibgram/domain/model/user"
)

type CommunityRequestStatus string

const (
	CommunityRequestStatusPending  CommunityRequestStatus = "PENDING"
	CommunityRequestStatusApproved CommunityRequestStatus = "APPROVED"
	CommunityRequestStatusDeclined CommunityRequestStatus = "DECLINED"
)

type CommunityRequest struct {
	CommunityID uint                     `gorm:"primaryKey;autoIncrement:false"`
	UserID      uint                     `gorm:"primaryKey;autoIncrement:false"`
	Community   communitymodel.Community `gorm:"foreignKey:CommunityID;references:CommunityID"`
	User        usermodel.User           `gorm:"foreignKey:UserID;references:UserID"`
	Status      CommunityRequestStatus   `gorm:"type:varchar(10);not null;default:'PENDING'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
