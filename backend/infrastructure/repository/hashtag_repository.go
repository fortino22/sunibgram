package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type hashtagRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

type trendingHashtagRow struct {
	Name        string
	ThreadCount int64
	Count24h    int64
	Count7d     int64
	Score       float64
}

func NewHashtagRepository(writeDB, readDB *gorm.DB) domainrepo.IHashtagRepository {
	return &hashtagRepository{writeDB: writeDB, readDB: readDB}
}

func (r *hashtagRepository) UpsertHashtags(ctx context.Context, names []string) (map[string]uuid.UUID, error) {
	ids := make(map[string]uuid.UUID, len(names))
	if len(names) == 0 {
		return ids, nil
	}

	hashtags := make([]model.Hashtag, len(names))
	for i, name := range names {
		hashtags[i] = model.Hashtag{Name: name}
	}
	if err := r.writeDB.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(&hashtags).Error; err != nil {
		return nil, err
	}

	var rows []model.Hashtag
	if err := r.writeDB.WithContext(ctx).
		Where("name IN ?", names).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		ids[row.Name] = row.ID
	}
	return ids, nil
}

func (r *hashtagRepository) LinkHashtagsToThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error {
	if len(hashtagIDs) == 0 {
		return nil
	}
	return r.writeDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inserted, err := insertThreadHashtagLinks(tx, threadID, hashtagIDs)
		if err != nil {
			return err
		}
		return patchHashtagThreadCounts(tx, inserted, 1)
	})
}

func (r *hashtagRepository) ReplaceHashtagsForThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error {
	return r.writeDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []uuid.UUID
		if err := tx.Model(&model.ThreadHashtag{}).
			Where("thread_id = ?", threadID).
			Pluck("hashtag_id", &existing).Error; err != nil {
			return err
		}

		existingSet := uuidSet(existing)
		nextSet := uuidSet(hashtagIDs)
		toRemove := make([]uuid.UUID, 0)
		for _, id := range existing {
			if _, ok := nextSet[id]; !ok {
				toRemove = append(toRemove, id)
			}
		}
		toAdd := make([]uuid.UUID, 0)
		for _, id := range hashtagIDs {
			if _, ok := existingSet[id]; !ok {
				toAdd = append(toAdd, id)
			}
		}

		if len(toRemove) > 0 {
			if err := tx.Where("thread_id = ? AND hashtag_id IN ?", threadID, toRemove).Delete(&model.ThreadHashtag{}).Error; err != nil {
				return err
			}
			if err := patchHashtagThreadCounts(tx, toRemove, -1); err != nil {
				return err
			}
		}
		inserted, err := insertThreadHashtagLinks(tx, threadID, toAdd)
		if err != nil {
			return err
		}
		return patchHashtagThreadCounts(tx, inserted, 1)
	})
}

func (r *hashtagRepository) UnlinkHashtagsFromThread(ctx context.Context, threadID uint) error {
	return r.writeDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []uuid.UUID
		if err := tx.Model(&model.ThreadHashtag{}).
			Where("thread_id = ?", threadID).
			Pluck("hashtag_id", &existing).Error; err != nil {
			return err
		}
		if len(existing) == 0 {
			return nil
		}
		if err := tx.Where("thread_id = ?", threadID).Delete(&model.ThreadHashtag{}).Error; err != nil {
			return err
		}
		return patchHashtagThreadCounts(tx, existing, -1)
	})
}

func (r *hashtagRepository) GetTrending(ctx context.Context, limit int, since24h, since7d time.Time) ([]model.TrendingHashtag, error) {
	var rows []trendingHashtagRow
	if err := r.readDB.WithContext(ctx).
		Table("hashtags").
		Select(`
			hashtags.name,
			hashtags.thread_count,
			COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END) AS count24h,
			COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END) AS count7d,
			(COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END)::float + 0.5 * COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END)::float) AS score
		`, since24h, since7d, since24h, since7d).
		Joins("JOIN thread_hashtags ON thread_hashtags.hashtag_id = hashtags.id").
		Joins("JOIN threads ON threads.thread_id = thread_hashtags.thread_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.created_at >= ?", since7d).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Group("hashtags.id, hashtags.name, hashtags.thread_count").
		Order("score DESC, count24h DESC, hashtags.name ASC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]model.TrendingHashtag, len(rows))
	for i, row := range rows {
		result[i] = model.TrendingHashtag{Name: row.Name, ThreadCount: row.ThreadCount, Score: row.Score}
	}
	return result, nil
}

func insertThreadHashtagLinks(tx *gorm.DB, threadID uint, hashtagIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(hashtagIDs) == 0 {
		return nil, nil
	}

	inserted := make([]uuid.UUID, 0, len(hashtagIDs))
	for _, hashtagID := range hashtagIDs {
		result := tx.Exec(`
			INSERT INTO thread_hashtags (thread_id, hashtag_id, created_at)
			VALUES (?, ?, NOW())
			ON CONFLICT DO NOTHING
		`, threadID, hashtagID)
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected > 0 {
			inserted = append(inserted, hashtagID)
		}
	}
	return inserted, nil
}

func patchHashtagThreadCounts(tx *gorm.DB, hashtagIDs []uuid.UUID, delta int) error {
	if len(hashtagIDs) == 0 {
		return nil
	}
	return tx.Model(&model.Hashtag{}).
		Where("id IN ?", hashtagIDs).
		UpdateColumn("thread_count", gorm.Expr("GREATEST(thread_count + ?, 0)", delta)).Error
}

func uuidSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
