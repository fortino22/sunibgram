package model

import "time"

type ChatReaction struct {
	ReactionID uint      `gorm:"primaryKey;autoIncrement"`
	ChatID     uint      `gorm:"not null;uniqueIndex:idx_chat_reaction_unique,priority:1"`
	UserID     uint      `gorm:"not null;uniqueIndex:idx_chat_reaction_unique,priority:2"`
	Emoji      string    `gorm:"type:varchar(32);not null;uniqueIndex:idx_chat_reaction_unique,priority:3"`
	CreatedAt  time.Time
}
