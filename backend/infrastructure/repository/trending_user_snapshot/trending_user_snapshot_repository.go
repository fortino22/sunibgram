package trendingusersnapshotrepo

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
	trendingusersnapshotrepoiface "github.com/sunibgram/domain/repository/trending_user_snapshot"
)

type trendingUserSnapshotRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewTrendingUserSnapshotRepository(writeDB, readDB *gorm.DB) trendingusersnapshotrepoiface.ITrendingUserSnapshotRepository {
	return &trendingUserSnapshotRepository{writeDB: writeDB, readDB: readDB}
}

func (r *trendingUserSnapshotRepository) UpsertWeek(ctx context.Context, weekStart time.Time, rows []trendingusersnapshotmodel.TrendingUserSnapshot) error {
	if len(rows) == 0 {
		return nil
	}
	weekDate := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	snapshotAt := time.Now().UTC()
	for i := range rows {
		rows[i].WeekStartDate = weekDate
		rows[i].SnapshotAt = snapshotAt
	}
	return r.writeDB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "week_start_date"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"rank", "score",
				"likes_in", "comments_in", "bookmarks_in", "reposts_in", "thread_count",
				"likes_out", "comments_out", "bookmarks_out", "posts_out",
				"snapshot_at",
			}),
		}).
		Create(&rows).Error
}

func (r *trendingUserSnapshotRepository) FindByUser(ctx context.Context, userID uint, limit int) ([]trendingusersnapshotmodel.TrendingUserSnapshot, error) {
	if limit < 1 || limit > 200 {
		limit = 30
	}
	var rows []trendingusersnapshotmodel.TrendingUserSnapshot
	err := r.readDB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("week_start_date DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *trendingUserSnapshotRepository) FindByUserAndMonth(ctx context.Context, userID uint, year, month int) ([]trendingusersnapshotmodel.TrendingUserSnapshot, error) {
	var rows []trendingusersnapshotmodel.TrendingUserSnapshot
	err := r.readDB.WithContext(ctx).
		Where("user_id = ? AND EXTRACT(YEAR FROM week_start_date) = ? AND EXTRACT(MONTH FROM week_start_date) = ?",
			userID, year, month).
		Order("week_start_date ASC").
		Find(&rows).Error
	return rows, err
}

func (r *trendingUserSnapshotRepository) FindAvailableMonths(ctx context.Context, userID uint) ([]trendingusersnapshotrepoiface.MonthBucket, error) {
	type row struct {
		Year  int
		Month int
	}
	var raw []row
	err := r.readDB.WithContext(ctx).
		Table("trending_user_snapshots").
		Select("DISTINCT EXTRACT(YEAR FROM week_start_date)::int AS year, EXTRACT(MONTH FROM week_start_date)::int AS month").
		Where("user_id = ?", userID).
		Order("year DESC, month DESC").
		Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	out := make([]trendingusersnapshotrepoiface.MonthBucket, len(raw))
	for i, r := range raw {
		out[i] = trendingusersnapshotrepoiface.MonthBucket{Year: r.Year, Month: r.Month}
	}
	return out, nil
}

func (r *trendingUserSnapshotRepository) FindByID(ctx context.Context, snapshotID uint) (*trendingusersnapshotmodel.TrendingUserSnapshot, error) {
	var row trendingusersnapshotmodel.TrendingUserSnapshot
	err := r.readDB.WithContext(ctx).
		Where("snapshot_id = ?", snapshotID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
