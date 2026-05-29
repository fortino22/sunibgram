package eventactionconfirmationmodel

import "time"

type ActionType string

const (
	ActionDeleteEvent ActionType = "DELETE_EVENT"
)

type EventActionConfirmation struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	EventID     string     `gorm:"type:varchar(64);index;not null"`
	ActorUserID uint       `gorm:"index;not null"`
	Type        ActionType `gorm:"type:varchar(32);not null;index"`
	EventTitle  string     `gorm:"type:varchar(200)"`
	CreatedAt   time.Time
}

func (EventActionConfirmation) TableName() string {
	return "event_action_confirmations"
}
