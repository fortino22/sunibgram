package eventparticipantmodel

import "time"

type EventParticipant struct {
	UserID   uint   `gorm:"primaryKey;autoIncrement:false;index:idx_event_participants_user,priority:1"`
	EventID  string `gorm:"type:varchar(36);primaryKey;index:idx_event_participants_user,priority:2"`
	JoinedAt time.Time
}
