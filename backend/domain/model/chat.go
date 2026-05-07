package model

import "time"

type Chat struct {
	ChatID    uint       `gorm:"primaryKey;autoIncrement"`
	RoomID    uint       `gorm:"not null;index;index:idx_chat_room_deleted_created,priority:1"`
	Room      ChatRoom   `gorm:"foreignKey:RoomID;references:RoomID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	SenderID  uint       `gorm:"not null"`
	Sender    User       `gorm:"foreignKey:SenderID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	Message   *string    `gorm:"type:text"`
	IsRead    bool       `gorm:"not null;default:false"`
	DeletedAt *time.Time `gorm:"index;index:idx_chat_room_deleted_created,priority:2"`
	CreatedAt time.Time  `gorm:"index:idx_chat_room_deleted_created,priority:3,sort:desc"`
}
