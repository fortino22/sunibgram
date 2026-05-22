package model

import "time"

type UserEventPin struct {
	UserID    uint   `gorm:"primaryKey;autoIncrement:false;index:idx_user_event_pins_user,priority:1"`
	EventID   string `gorm:"type:varchar(36);primaryKey;index:idx_user_event_pins_user,priority:2"`
	CreatedAt time.Time
}
