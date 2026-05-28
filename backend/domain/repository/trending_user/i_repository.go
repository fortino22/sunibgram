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

type ITrendingUserRepository interface {
	GetWeeklyTrending(ctx context.Context, viewerID uint, page, limit int) (TrendingUsersPage, error)
}
