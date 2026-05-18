package model

import "time"

type ChatRoomType string

const (
	ChatRoomTypePrivate ChatRoomType = "private"
	ChatRoomTypeGroup   ChatRoomType = "group"
)

type ChatRoom struct {
	RoomID       uint         `gorm:"primaryKey;autoIncrement"`
	Name         *string      `gorm:"type:varchar(100)"`
	Description  *string      `gorm:"type:text"`
	AvatarURL    *string      `gorm:"type:varchar(500)"`
	Type         ChatRoomType `gorm:"type:varchar(10);not null"`
	MemberCount  int          `gorm:"not null;default:0"`
	EncryptedKey *string      `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
