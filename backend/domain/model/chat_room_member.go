package model

import "time"

type ChatRoomMember struct {
	RoomID   uint     `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_chat_room_member,priority:1"`
	Room     ChatRoom `gorm:"foreignKey:RoomID;references:RoomID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	UserID   uint     `gorm:"primaryKey;autoIncrement:false;uniqueIndex:idx_chat_room_member,priority:2"`
	User     User     `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	JoinedAt time.Time
	IsMuted  bool `gorm:"not null;default:false"`
}
