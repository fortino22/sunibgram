package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type interactionTimestamp struct {
	ThreadID     uint
	InteractedAt time.Time
}

type interactionRepo struct {
	writeDB  *gorm.DB
	readDB   *gorm.DB
	cache    *cache.Cache
	table    string
	ttl      time.Duration
	threadDB *threadDB
}

func newInteractionRepo(writeDB, readDB *gorm.DB, c *cache.Cache, table string, tdb *threadDB) *interactionRepo {
	return &interactionRepo{
		writeDB:  writeDB,
		readDB:   readDB,
		cache:    c,
		table:    table,
		ttl:      cache.DefaultTTL,
		threadDB: tdb,
	}
}

func (r *interactionRepo) listKey(userID uint, page, limit int) string {
	return fmt.Sprintf("%s:user:%d:page:%d:limit:%d", r.table, userID, page, limit)
}

// interactionThreadTag returns the reverse-index tag for cache keys whose
// invalidation is keyed on (table, threadID). Shared with event_repository.go
// via interactionThreadCacheKeysKey — kept symmetric so a key registered here
// is invalidatable from there.
func (r *interactionRepo) interactionThreadTag(threadID uint) string {
	return interactionThreadCacheKeysKey(r.table, threadID)
}

func (r *interactionRepo) invalidateUser(ctx context.Context, userID uint) {
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("%s:user:%d:*", r.table, userID))
}

func (r *interactionRepo) invalidateByThreadID(ctx context.Context, threadID uint) {
	_ = r.cache.InvalidateTag(ctx, r.interactionThreadTag(threadID))
}

func (r *interactionRepo) deleteByThreadID(threadID uint) error {
	return r.writeDB.Table(r.table).Where("thread_id = ?", threadID).Delete(nil).Error
}

func (r *interactionRepo) saveRecord(record any) error {
	return r.writeDB.Create(record).Error
}

func (r *interactionRepo) deleteRecord(userID, threadID uint, model any) error {
	return r.writeDB.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(model).Error
}

func (r *interactionRepo) exists(userID, threadID uint, model any) (bool, error) {
	var count int64
	err := r.readDB.Model(model).Where("user_id = ? AND thread_id = ?", userID, threadID).Count(&count).Error
	return count > 0, err
}

func (r *interactionRepo) areThreadsInteracted(userID uint, threadIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(threadIDs))
	if len(threadIDs) == 0 {
		return result, nil
	}

	var ids []uint
	err := r.readDB.Table(r.table).
		Select("thread_id").
		Where("user_id = ? AND thread_id IN ?", userID, threadIDs).
		Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

// allUserThreadIDs returns every thread_id the user has interacted with for this table.
// Used to hydrate Redis viewer-flag sets.
func (r *interactionRepo) allUserThreadIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.readDB.Table(r.table).
		Select("thread_id").
		Where("user_id = ?", userID).
		Scan(&ids).Error
	return ids, err
}

func (r *interactionRepo) findByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	type cachedResult struct {
		Threads    []model.Thread
		Timestamps []time.Time
		Total      int64
	}

	key := r.listKey(userID, page, limit)
	var result cachedResult
	if err := r.cache.Get(context.Background(), key, &result); err == nil {
		return result.Threads, result.Timestamps, result.Total, nil
	}

	var total int64
	r.readDB.Table(r.table).Where("user_id = ?", userID).Count(&total)

	var rows []interactionTimestamp
	offset := (page - 1) * limit
	err := r.readDB.Table(r.table).
		Select("thread_id, created_at AS interacted_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, nil, total, err
	}

	ids := make([]uint, len(rows))
	orderMap := make(map[uint]time.Time, len(rows))
	for i, row := range rows {
		ids[i] = row.ThreadID
		orderMap[row.ThreadID] = row.InteractedAt
	}

	var threads []model.Thread
	if err := preloadThreadDetails(r.readDB).Where("thread_id IN ?", ids).Find(&threads).Error; err != nil {
		return nil, nil, total, err
	}

	threadMap := make(map[uint]model.Thread, len(threads))
	for _, t := range threads {
		threadMap[t.ThreadID] = t
	}
	sorted := make([]model.Thread, len(ids))
	timestamps := make([]time.Time, len(ids))
	for i, id := range ids {
		sorted[i] = threadMap[id]
		timestamps[i] = orderMap[id]
	}

	ctx := context.Background()
	if err := r.cache.Set(ctx, key, cachedResult{
		Threads:    sorted,
		Timestamps: timestamps,
		Total:      total,
	}, r.ttl); err == nil {
		for _, threadID := range ids {
			if threadID != 0 {
				r.cache.RegisterCacheKey(ctx, r.interactionThreadTag(threadID), key)
			}
		}
	}

	return sorted, timestamps, total, nil
}
