package model

import "time"

type Chat struct {
	ChatID    uint       `gorm:"primaryKey;autoIncrement"`
	RoomID    uint       `gorm:"not null;index"`
	Room      ChatRoom   `gorm:"foreignKey:RoomID;references:RoomID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	SenderID  uint       `gorm:"not null"`
	Sender    User       `gorm:"foreignKey:SenderID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	Message   *string    `gorm:"type:text"`
	IsRead    bool       `gorm:"not null;default:false"`
	DeletedAt *time.Time `gorm:"index"`
	CreatedAt time.Time
}
