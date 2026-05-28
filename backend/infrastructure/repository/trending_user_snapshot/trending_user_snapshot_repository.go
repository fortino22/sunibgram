package trendingusersnapshotrepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
	trendingusersnapshotrepoiface "github.com/sunibgram/domain/repository/trending_user_snapshot"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

const (
	snapshotByIDTTL   = 24 * time.Hour
	userHistoryTTL    = 6 * time.Hour
	availableMonthsTTL = 6 * time.Hour
)

type trendingUserSnapshotRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewTrendingUserSnapshotRepository(writeDB, readDB *gorm.DB, c *cache.Cache) trendingusersnapshotrepoiface.ITrendingUserSnapshotRepository {
	return &trendingUserSnapshotRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func snapshotByIDKey(snapshotID uint) string {
	return fmt.Sprintf("trending:snapshot:%d", snapshotID)
}

func userHistoryRangeKey(userID uint, start, end time.Time) string {
	return fmt.Sprintf("trending:user-history:%d:%s:%s",
		userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
}

func availableMonthsKey(userID uint) string {
	return fmt.Sprintf("trending:available-months:%d", userID)
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
	if err := r.writeDB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "week_start_date"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"rank", "score",
				"likes_in", "comments_in", "bookmarks_in", "reposts_in", "thread_count",
				"likes_out", "comments_out", "bookmarks_out", "posts_out",
				"snapshot_at",
			}),
		}).
		Create(&rows).Error; err != nil {
		return err
	}

	// Weekly snapshot batch landed. Drop all cached history pages + snapshot
	// rows so readers see fresh ranks/scores. Per-user tags catch any caches
	// for users in this batch; global tag catches anything missed.
	if r.cache != nil {
		tags := make([]string, 0, len(rows)+1)
		tags = append(tags, infrarepoShared.TrendingHistoryGlobalTag())
		for _, row := range rows {
			tags = append(tags, infrarepoShared.TrendingUserHistoryTag(row.UserID))
		}
		_ = r.cache.InvalidateTags(ctx, tags...)
	}
	return nil
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

func (r *trendingUserSnapshotRepository) FindByUserAndRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]trendingusersnapshotmodel.TrendingUserSnapshot, error) {
	key := userHistoryRangeKey(userID, startDate, endDate)
	if r.cache != nil {
		var cached []trendingusersnapshotmodel.TrendingUserSnapshot
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}

	var rows []trendingusersnapshotmodel.TrendingUserSnapshot
	if err := r.readDB.WithContext(ctx).
		Where("user_id = ? AND week_start_date >= ? AND week_start_date <= ?",
			userID, startDate, endDate).
		Order("week_start_date ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, key, rows, userHistoryTTL)
		_ = r.cache.RegisterCacheKeyMany(ctx, []string{
			infrarepoShared.TrendingUserHistoryTag(userID),
			infrarepoShared.TrendingHistoryGlobalTag(),
		}, key)
	}
	return rows, nil
}

func (r *trendingUserSnapshotRepository) FindAvailableMonths(ctx context.Context, userID uint) ([]trendingusersnapshotrepoiface.MonthBucket, error) {
	key := availableMonthsKey(userID)
	if r.cache != nil {
		var cached []trendingusersnapshotrepoiface.MonthBucket
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}

	type row struct {
		Year  int
		Month int
	}
	var raw []row
	if err := r.readDB.WithContext(ctx).
		Table("trending_user_snapshots").
		Select("DISTINCT EXTRACT(YEAR FROM week_start_date)::int AS year, EXTRACT(MONTH FROM week_start_date)::int AS month").
		Where("user_id = ?", userID).
		Order("year DESC, month DESC").
		Scan(&raw).Error; err != nil {
		return nil, err
	}
	out := make([]trendingusersnapshotrepoiface.MonthBucket, len(raw))
	for i, r := range raw {
		out[i] = trendingusersnapshotrepoiface.MonthBucket{Year: r.Year, Month: r.Month}
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, key, out, availableMonthsTTL)
		_ = r.cache.RegisterCacheKeyMany(ctx, []string{
			infrarepoShared.TrendingUserHistoryTag(userID),
			infrarepoShared.TrendingHistoryGlobalTag(),
		}, key)
	}
	return out, nil
}

func (r *trendingUserSnapshotRepository) FindByID(ctx context.Context, snapshotID uint) (*trendingusersnapshotmodel.TrendingUserSnapshot, error) {
	key := snapshotByIDKey(snapshotID)
	if r.cache != nil {
		var cached trendingusersnapshotmodel.TrendingUserSnapshot
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return &cached, nil
		}
	}

	var row trendingusersnapshotmodel.TrendingUserSnapshot
	if err := r.readDB.WithContext(ctx).
		Where("snapshot_id = ?", snapshotID).
		First(&row).Error; err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, key, row, snapshotByIDTTL)
		_ = r.cache.RegisterCacheKeyMany(ctx, []string{
			infrarepoShared.TrendingUserHistoryTag(row.UserID),
			infrarepoShared.TrendingHistoryGlobalTag(),
		}, key)
	}
	return &row, nil
}
