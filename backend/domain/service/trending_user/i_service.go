package trendingusersvciface

import (
	"context"
	"time"

	exploremodel "github.com/sunibgram/domain/model/explore"
	trendingusermodel "github.com/sunibgram/domain/model/trending_user"
	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
	trendingusersnapshotrepoiface "github.com/sunibgram/domain/repository/trending_user_snapshot"
)

type MyCurrentResult struct {
	User *trendingusermodel.TrendingUser
	Rank *int
	Page trendinguserrepoiface.TrendingUsersPage
}

type MyHistoryResult struct {
	Rows            []trendingusersnapshotmodel.TrendingUserSnapshot
	StartDate       time.Time
	EndDate         time.Time
	AvailableMonths []trendingusersnapshotrepoiface.MonthBucket
}

type ITrendingUserService interface {
	GetWeeklyTrending(ctx context.Context, viewerID uint, page, limit int) (trendinguserrepoiface.TrendingUsersPage, error)
	GetMyCurrent(ctx context.Context, viewerID uint) (MyCurrentResult, error)
	GetMyHistory(ctx context.Context, viewerID uint, startDate, endDate *time.Time) (MyHistoryResult, error)
	GetSnapshotByID(ctx context.Context, viewerID, snapshotID uint) (*trendingusersnapshotmodel.TrendingUserSnapshot, error)
	SearchMutualUsers(ctx context.Context, viewerID uint, query string, limit int) ([]exploremodel.ExploreUserResult, error)
	GetUserHistory(ctx context.Context, viewerID, targetID uint, startDate, endDate *time.Time) (MyHistoryResult, error)
}
