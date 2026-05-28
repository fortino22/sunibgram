package trendingusersnapshotmodel

import "time"

type TrendingUserSnapshot struct {
	SnapshotID    uint      `gorm:"primaryKey;autoIncrement"`
	WeekStartDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_snapshot_week_user,priority:1;index:idx_snapshot_week"`
	UserID        uint      `gorm:"not null;uniqueIndex:idx_snapshot_week_user,priority:2;index:idx_snapshot_user_week"`
	Rank          int       `gorm:"not null"`
	Score         float64   `gorm:"not null"`
	LikesIn       int       `gorm:"not null;default:0"`
	CommentsIn    int       `gorm:"not null;default:0"`
	BookmarksIn   int       `gorm:"not null;default:0"`
	RepostsIn     int       `gorm:"not null;default:0"`
	ThreadCount   int       `gorm:"not null;default:0"`
	LikesOut      int       `gorm:"not null;default:0"`
	CommentsOut   int       `gorm:"not null;default:0"`
	BookmarksOut  int       `gorm:"not null;default:0"`
	PostsOut      int       `gorm:"not null;default:0"`
	SnapshotAt    time.Time `gorm:"type:timestamptz;not null"`
}
