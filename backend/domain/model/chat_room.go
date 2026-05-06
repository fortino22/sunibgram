package model

import "time"

type ChatRoomType string

const (
	ChatRoomTypePrivate ChatRoomType = "private"
	ChatRoomTypeGroup   ChatRoomType = "group"
)

type ChatRoom struct {
	RoomID    uint         `gorm:"primaryKey;autoIncrement"`
	Name      *string      `gorm:"type:varchar(100)"`
	Type      ChatRoomType `gorm:"type:varchar(10);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
