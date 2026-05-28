package trendingusersnapshotrepoiface

import (
	"context"
	"time"

	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
)

type MonthBucket struct {
	Year  int
	Month int
}

type ITrendingUserSnapshotRepository interface {
	UpsertWeek(ctx context.Context, weekStart time.Time, rows []trendingusersnapshotmodel.TrendingUserSnapshot) error
	FindByUser(ctx context.Context, userID uint, limit int) ([]trendingusersnapshotmodel.TrendingUserSnapshot, error)
	FindByUserAndMonth(ctx context.Context, userID uint, year, month int) ([]trendingusersnapshotmodel.TrendingUserSnapshot, error)
	FindAvailableMonths(ctx context.Context, userID uint) ([]MonthBucket, error)
	FindByID(ctx context.Context, snapshotID uint) (*trendingusersnapshotmodel.TrendingUserSnapshot, error)
}
