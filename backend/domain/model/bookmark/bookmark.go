package bookmarkmodel

import "time"

type Bookmark struct {
	UserID    uint `gorm:"primaryKey;autoIncrement:false;index:idx_bookmarks_user_thread,priority:1"`
	ThreadID  uint `gorm:"primaryKey;autoIncrement:false;index:idx_bookmarks_user_thread,priority:2"`
	CreatedAt time.Time
}
