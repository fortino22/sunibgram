package model

import "time"

type ChatRoomMember struct {
	RoomID   uint     `gorm:"primaryKey;autoIncrement:false"`
	Room     ChatRoom `gorm:"foreignKey:RoomID;references:RoomID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	UserID   uint     `gorm:"primaryKey;autoIncrement:false"`
	User     User     `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	JoinedAt time.Time
}
