package eventthreadpinmodel

import "time"

type EventThreadPin struct {
	EventID   string `gorm:"type:varchar(36);primaryKey;column:event_id"`
	ThreadID  uint   `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
}
