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
	UserID           uint       `gorm:"not null;index;index:idx_thread_user_type,priority:1"`
	User             User       `gorm:"foreignKey:UserID;references:UserID"`
	CommunityID      *uint      `gorm:"index;index:idx_thread_feed,priority:2"`
	Community        *Community `gorm:"foreignKey:CommunityID;references:CommunityID"`
	Type             ThreadType `gorm:"type:varchar(10);not null;default:'POST';index;index:idx_thread_user_type,priority:2;index:idx_thread_feed,priority:1"`
	Content          string     `gorm:"type:text;not null"`
	ParentID         *uint      `gorm:"index"`
	RootID           *uint      `gorm:"index"`
	OriginalThreadID *uint      `gorm:"index"`
	OriginalThread   *Thread    `gorm:"foreignKey:OriginalThreadID;references:ThreadID"`
	LikeCount        int        `gorm:"not null;default:0"`
	RepostCount      int        `gorm:"not null;default:0"`
	BookmarkCount    int        `gorm:"not null;default:0"`
	CommentCount     int        `gorm:"not null;default:0"`
	FeedScore        float64    `gorm:"column:feed_score;default:0"`
	CreatedAt        time.Time  `gorm:"index:idx_thread_user_type,priority:3,sort:desc;index:idx_thread_feed,priority:3,sort:desc"`
	UpdatedAt        time.Time
	ThreadEvents     []ThreadEvent `gorm:"foreignKey:ThreadID;references:ThreadID"`
	EventID          *string       `gorm:"-"`
	EventIDUpdated   bool          `gorm:"-"`
}
