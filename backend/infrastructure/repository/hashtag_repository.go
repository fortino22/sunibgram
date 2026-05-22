package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
)

type hashtagRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
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

func (r *hashtagRepository) GetTrending(ctx context.Context, limit int) ([]model.TrendingHashtag, error) {
	var rows []model.Hashtag
	if err := r.readDB.WithContext(ctx).
		Where("week_thread_count > 0 OR week_like_count > 0 OR week_repost_count > 0").
		Order("(week_thread_count * 3 + week_like_count + week_repost_count * 2) DESC, thread_count DESC, name ASC").
		Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]model.TrendingHashtag, len(rows))
	for i, row := range rows {
		result[i] = model.TrendingHashtag{
			Name:        row.Name,
			ThreadCount: row.ThreadCount,
		}
	}
	return result, nil
}

func (r *hashtagRepository) PatchHashtagCounterByThreadID(ctx context.Context, threadID uint, column string, delta int) error {
	if delta == 0 {
		return nil
	}

	var hashtagIDs []uuid.UUID
	if err := r.writeDB.WithContext(ctx).Model(&model.ThreadHashtag{}).
		Where("thread_id = ?", threadID).
		Pluck("hashtag_id", &hashtagIDs).Error; err != nil {
		return err
	}
	if len(hashtagIDs) == 0 {
		var thread model.Thread
		if err := r.writeDB.WithContext(ctx).Select("root_id", "parent_id", "original_thread_id").Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
			return err
		}
		if thread.RootID != nil {
			threadID = *thread.RootID
		} else if thread.ParentID != nil {
			threadID = *thread.ParentID
		} else if thread.OriginalThreadID != nil {
			threadID = *thread.OriginalThreadID
		} else {
			return nil
		}
		if err := r.writeDB.WithContext(ctx).Model(&model.ThreadHashtag{}).
			Where("thread_id = ?", threadID).
			Pluck("hashtag_id", &hashtagIDs).Error; err != nil {
			return err
		}
		if len(hashtagIDs) == 0 {
			return nil
		}
	}

	updates := map[string]any{
		column: gorm.Expr("GREATEST("+column+" + ?, 0)", delta),
	}
	switch column {
	case "thread_count":
		updates["week_thread_count"] = gorm.Expr("GREATEST(week_thread_count + ?, 0)", delta)
	case "like_count":
		updates["week_like_count"] = gorm.Expr("GREATEST(week_like_count + ?, 0)", delta)
	case "repost_count":
		updates["week_repost_count"] = gorm.Expr("GREATEST(week_repost_count + ?, 0)", delta)
	case "week_thread_count", "week_like_count", "week_repost_count":
	default:
		return nil
	}

	return r.writeDB.WithContext(ctx).Model(&model.Hashtag{}).
		Where("id IN ?", hashtagIDs).
		UpdateColumns(updates).Error
}

func insertThreadHashtagLinks(tx *gorm.DB, threadID uint, hashtagIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(hashtagIDs) == 0 {
		return nil, nil
	}

	var query strings.Builder
	query.WriteString("INSERT INTO thread_hashtags (thread_id, hashtag_id, created_at) VALUES ")
	args := make([]any, 0, len(hashtagIDs)*2)
	for i, hashtagID := range hashtagIDs {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString("(?, ?, NOW())")
		args = append(args, threadID, hashtagID)
	}
	query.WriteString(" ON CONFLICT DO NOTHING RETURNING hashtag_id")

	inserted := make([]uuid.UUID, 0, len(hashtagIDs))
	if err := tx.Raw(query.String(), args...).Scan(&inserted).Error; err != nil {
		return nil, err
	}
	return inserted, nil
}

func patchHashtagThreadCounts(tx *gorm.DB, hashtagIDs []uuid.UUID, delta int) error {
	if len(hashtagIDs) == 0 {
		return nil
	}
	return tx.Model(&model.Hashtag{}).
		Where("id IN ?", hashtagIDs).
		UpdateColumns(map[string]any{
			"thread_count":      gorm.Expr("GREATEST(thread_count + ?, 0)", delta),
			"week_thread_count": gorm.Expr("GREATEST(week_thread_count + ?, 0)", delta),
		}).Error
}

func patchHashtagCounters(db *gorm.DB, c *cache.Cache, hashtagIDs []uuid.UUID, column string, delta int) {
	if len(hashtagIDs) == 0 || delta == 0 {
		return
	}
	updates := map[string]any{
		column: gorm.Expr("GREATEST("+column+" + ?, 0)", delta),
	}
	switch column {
	case "thread_count":
		updates["week_thread_count"] = gorm.Expr("GREATEST(week_thread_count + ?, 0)", delta)
	case "like_count":
		updates["week_like_count"] = gorm.Expr("GREATEST(week_like_count + ?, 0)", delta)
	case "repost_count":
		updates["week_repost_count"] = gorm.Expr("GREATEST(week_repost_count + ?, 0)", delta)
	default:
		return
	}
	db.Model(&model.Hashtag{}).
		Where("id IN ?", hashtagIDs).
		UpdateColumns(updates)
	if c != nil {
		_ = c.InvalidateTag(context.Background(), trendingHashtagsTag())
	}
}

func patchHashtagCounterByThreadID(db *gorm.DB, c *cache.Cache, threadID uint, column string, delta int) {
	var hashtagIDs []uuid.UUID
	if err := db.Model(&model.ThreadHashtag{}).
		Where("thread_id = ?", threadID).
		Pluck("hashtag_id", &hashtagIDs).Error; err != nil {
		return
	}
	if len(hashtagIDs) == 0 {
		var thread model.Thread
		if err := db.Select("root_id", "parent_id", "original_thread_id").Where("thread_id = ?", threadID).First(&thread).Error; err != nil {
			return
		}
		if thread.RootID != nil {
			threadID = *thread.RootID
		} else if thread.ParentID != nil {
			threadID = *thread.ParentID
		} else if thread.OriginalThreadID != nil {
			threadID = *thread.OriginalThreadID
		} else {
			return
		}
		if err := db.Model(&model.ThreadHashtag{}).
			Where("thread_id = ?", threadID).
			Pluck("hashtag_id", &hashtagIDs).Error; err != nil {
			return
		}
		if len(hashtagIDs) == 0 {
			return
		}
	}
	patchHashtagCounters(db, c, hashtagIDs, column, delta)
}

func uuidSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
