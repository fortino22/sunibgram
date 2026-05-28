package trendinguserrepoiface

import (
	"context"
	"time"

	trendingusermodel "github.com/sunibgram/domain/model/trending_user"
)

type TrendingUsersPage struct {
	Users     []trendingusermodel.TrendingUser
	Total     int64
	CachedAt  time.Time
	RefreshAt time.Time
}

type WeeklyScoreRow struct {
	UserID       uint
	Score        float64
	ThreadCount  int
	LikesIn      int
	CommentsIn   int
	BookmarksIn  int
	RepostsIn    int
	LikesOut     int
	CommentsOut  int
	BookmarksOut int
	PostsOut     int
}

type ITrendingUserRepository interface {
	GetWeeklyTrending(ctx context.Context, viewerID uint, page, limit int) (TrendingUsersPage, error)
	GetMyWeeklyTrending(ctx context.Context, viewerID uint) (user *trendingusermodel.TrendingUser, rank *int, page TrendingUsersPage, err error)
	ComputeAllWeeklyScores(ctx context.Context) ([]WeeklyScoreRow, error)
}
