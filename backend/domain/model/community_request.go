package model

import "time"

type CommunityRequestStatus string

const (
	CommunityRequestStatusPending  CommunityRequestStatus = "PENDING"
	CommunityRequestStatusApproved CommunityRequestStatus = "APPROVED"
	CommunityRequestStatusDeclined CommunityRequestStatus = "DECLINED"
)

type CommunityRequest struct {
	CommunityID uint                   `gorm:"primaryKey;autoIncrement:false"`
	UserID      uint                   `gorm:"primaryKey;autoIncrement:false"`
	Community   Community              `gorm:"foreignKey:CommunityID;references:CommunityID"`
	User        User                   `gorm:"foreignKey:UserID;references:UserID"`
	Status      CommunityRequestStatus `gorm:"type:varchar(10);not null;default:'PENDING'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
