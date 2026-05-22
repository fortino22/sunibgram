package model

import "time"

type ChatRead struct {
	ReadID uint `gorm:"primaryKey;autoIncrement"`
	ChatID uint `gorm:"not null;index;uniqueIndex:idx_chat_reads_chat_user"`
	Chat   Chat `gorm:"foreignKey:ChatID;references:ChatID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE"`
	UserID uint `gorm:"not null;index;uniqueIndex:idx_chat_reads_chat_user"`
	ReadAt time.Time
}

func (ChatRead) TableName() string {
	return "chat_reads"
}
