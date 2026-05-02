package model

import "time"

type ThreadType string

const (
	ThreadTypePost   ThreadType = "POST"
	ThreadTypeReply  ThreadType = "REPLY"
	ThreadTypeRepost ThreadType = "REPOST"
)

type Thread struct {
	ThreadID         uint       `gorm:"primaryKey;autoIncrement"`
	UserID           uint       `gorm:"not null;index"`
	User             User       `gorm:"foreignKey:UserID;references:UserID"`
	Type             ThreadType `gorm:"type:varchar(10);not null;default:'POST';index"`
	Content          string     `gorm:"type:text;not null"`
	ParentID         *uint      `gorm:"index"`
	RootID           *uint      `gorm:"index"`
	OriginalThreadID *uint      `gorm:"index"`
	OriginalThread   *Thread    `gorm:"foreignKey:OriginalThreadID;references:ThreadID"`
	LikeCount        int        `gorm:"not null;default:0"`
	RepostCount      int        `gorm:"not null;default:0"`
	BookmarkCount    int        `gorm:"not null;default:0"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
