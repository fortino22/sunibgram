package communityactionconfirmationmodel

import "time"

type ActionType string

const (
	ActionDeleteCommunity ActionType = "DELETE_COMMUNITY"
	ActionAdminTransfer   ActionType = "ADMIN_TRANSFER"
)

type CommunityActionConfirmation struct {
	ID            uint       `gorm:"primaryKey;autoIncrement"`
	CommunityID   uint       `gorm:"index;not null"`
	ActorUserID   uint       `gorm:"index;not null"`
	TargetUserID  *uint      `gorm:"index"`
	Type          ActionType `gorm:"type:varchar(32);not null;index"`
	CommunityName string     `gorm:"type:varchar(120)"`
	Note          string     `gorm:"type:varchar(255)"`
	CreatedAt     time.Time
}

func (CommunityActionConfirmation) TableName() string {
	return "community_action_confirmations"
}
