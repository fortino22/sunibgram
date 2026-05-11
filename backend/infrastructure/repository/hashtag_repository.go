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
	Name     string
	Count24h int64
	Count7d  int64
	Score    float64
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

	links := make([]model.ThreadHashtag, len(hashtagIDs))
	for i, hashtagID := range hashtagIDs {
		links[i] = model.ThreadHashtag{ThreadID: threadID, HashtagID: hashtagID}
	}
	return r.writeDB.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&links).Error
}

func (r *hashtagRepository) ReplaceHashtagsForThread(ctx context.Context, threadID uint, hashtagIDs []uuid.UUID) error {
	return r.writeDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("thread_id = ?", threadID).Delete(&model.ThreadHashtag{}).Error; err != nil {
			return err
		}
		if len(hashtagIDs) == 0 {
			return nil
		}
		links := make([]model.ThreadHashtag, len(hashtagIDs))
		for i, hashtagID := range hashtagIDs {
			links[i] = model.ThreadHashtag{ThreadID: threadID, HashtagID: hashtagID}
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error
	})
}

func (r *hashtagRepository) GetTrending(ctx context.Context, limit int, since24h, since7d time.Time) ([]model.TrendingHashtag, error) {
	var rows []trendingHashtagRow
	if err := r.readDB.WithContext(ctx).
		Table("hashtags").
		Select(`
			hashtags.name,
			COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END) AS count24h,
			COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END) AS count7d,
			(COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END)::float + 0.5 * COUNT(DISTINCT CASE WHEN threads.created_at >= ? THEN threads.thread_id END)::float) AS score
		`, since24h, since7d, since24h, since7d).
		Joins("JOIN thread_hashtags ON thread_hashtags.hashtag_id = hashtags.id").
		Joins("JOIN threads ON threads.thread_id = thread_hashtags.thread_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Where("threads.created_at >= ?", since7d).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Group("hashtags.id, hashtags.name").
		Order("score DESC, count24h DESC, hashtags.name ASC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]model.TrendingHashtag, len(rows))
	for i, row := range rows {
		result[i] = model.TrendingHashtag{Name: row.Name, ThreadCount: row.Count7d, Score: row.Score}
	}
	return result, nil
}

func (r *hashtagRepository) GetThreadsByHashtag(ctx context.Context, name string, cursor *uint, userID uint, limit int) ([]model.Thread, *uint, error) {
	query := r.readDB.WithContext(ctx).
		Model(&model.Thread{}).
		Select("threads.*").
		Joins("JOIN thread_hashtags ON thread_hashtags.thread_id = threads.thread_id").
		Joins("JOIN hashtags ON hashtags.id = thread_hashtags.hashtag_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN user_settings ON user_settings.user_id = threads.user_id").
		Joins("LEFT JOIN user_relationships ON user_relationships.follower_id = ? AND user_relationships.following_id = threads.user_id", userID).
		Where("hashtags.name = ?", name).
		Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeReply)}).
		Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
		Where("threads.user_id = ? OR COALESCE(user_settings.is_private, false) = false OR user_relationships.following_id IS NOT NULL", userID)
	if cursor != nil {
		query = query.Where("threads.thread_id < ?", *cursor)
	}

	var rows []model.Thread
	if err := query.
		Preload("User", preloadThreadUser).
		Preload("Community").
		Preload("User.UserFaculty.Faculty").
		Order("threads.thread_id DESC").
		Limit(limit + 1).
		Find(&rows).Error; err != nil {
		return nil, nil, err
	}

	var nextCursor *uint
	if len(rows) > limit {
		next := rows[limit-1].ThreadID
		nextCursor = &next
		rows = rows[:limit]
	}
	return rows, nextCursor, nil
}
