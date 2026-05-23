package infrarepoShared

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	threadmodel "github.com/sunibgram/domain/model/thread"
	"github.com/sunibgram/infrastructure/cache"
)

type interactionTimestamp struct {
	ThreadID     uint
	InteractedAt time.Time
}

// PreloadFn is the contract InteractionRepo uses to load thread details after
// resolving thread IDs. Callers (likerepo, bookmarkrepo, repostrepo) pass
// `threadrepo.PreloadThreadDetails` at construction so InteractionRepo itself
// stays decoupled from the thread repository package — preventing an
// infrastructure/repository cycle.
type PreloadFn func(*gorm.DB) *gorm.DB

type InteractionRepo struct {
	WriteDB *gorm.DB
	ReadDB  *gorm.DB
	Cache   *cache.Cache
	table   string
	ttl     time.Duration
	preload PreloadFn
}

func NewInteractionRepo(WriteDB, ReadDB *gorm.DB, c *cache.Cache, table string, preload PreloadFn) *InteractionRepo {
	return &InteractionRepo{
		WriteDB: WriteDB,
		ReadDB:  ReadDB,
		Cache:   c,
		table:   table,
		ttl:     cache.DefaultTTL,
		preload: preload,
	}
}

func (r *InteractionRepo) ListKey(userID uint, page, limit int) string {
	return fmt.Sprintf("%s:user:%d:page:%d:limit:%d", r.table, userID, page, limit)
}

// InteractionThreadTag returns the reverse-index tag for cache keys whose
// invalidation is keyed on (table, threadID). Shared with event_repository.go
// via InteractionThreadCacheKeysKey — kept symmetric so a key registered here
// is invalidatable from there.
func (r *InteractionRepo) InteractionThreadTag(threadID uint) string {
	return InteractionThreadCacheKeysKey(r.table, threadID)
}

func (r *InteractionRepo) InvalidateUser(ctx context.Context, userID uint) {
	// InteractionUserListTag is registered at every FindByUserID Set site, so
	// InvalidateTag drops all pages of this user's interaction list in
	// O(set size) — no SCAN.
	_ = r.Cache.InvalidateTag(ctx, InteractionUserListTag(r.table, userID))
}

func (r *InteractionRepo) InvalidateByThreadID(ctx context.Context, threadID uint) {
	_ = r.Cache.InvalidateTag(ctx, r.InteractionThreadTag(threadID))
}

func (r *InteractionRepo) DeleteByThreadID(threadID uint) error {
	return r.WriteDB.Table(r.table).Where("thread_id = ?", threadID).Delete(nil).Error
}

func (r *InteractionRepo) SaveRecord(record any) error {
	return r.WriteDB.Create(record).Error
}

func (r *InteractionRepo) DeleteRecord(userID, threadID uint, model any) error {
	return r.WriteDB.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(model).Error
}

func (r *InteractionRepo) Exists(userID, threadID uint, model any) (bool, error) {
	var exists bool
	err := r.ReadDB.Model(model).Select("1").Where("user_id = ? AND thread_id = ?", userID, threadID).Limit(1).Scan(&exists).Error
	return exists, err
}

func (r *InteractionRepo) AreThreadsInteracted(userID uint, threadIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(threadIDs))
	if len(threadIDs) == 0 {
		return result, nil
	}

	var ids []uint
	err := r.ReadDB.Table(r.table).
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

// AllUserThreadIDs returns every thread_id the user has interacted with for this table.
// Used to hydrate Redis viewer-flag sets.
func (r *InteractionRepo) AllUserThreadIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.ReadDB.Table(r.table).
		Select("thread_id").
		Where("user_id = ?", userID).
		Scan(&ids).Error
	return ids, err
}

func (r *InteractionRepo) FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	type cachedResult struct {
		Threads    []threadmodel.Thread
		Timestamps []time.Time
		Total      int64
	}

	key := r.ListKey(userID, page, limit)
	var result cachedResult
	if err := r.Cache.Get(context.Background(), key, &result); err == nil {
		return result.Threads, result.Timestamps, result.Total, nil
	}

	var total int64
	r.ReadDB.Table(r.table).Where("user_id = ?", userID).Count(&total)

	var rows []interactionTimestamp
	offset := (page - 1) * limit
	err := r.ReadDB.Table(r.table).
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

	var threads []threadmodel.Thread
	if err := r.preload(r.ReadDB).Where("threads.thread_id IN ?", ids).Find(&threads).Error; err != nil {
		return nil, nil, total, err
	}

	threadMap := make(map[uint]threadmodel.Thread, len(threads))
	for _, t := range threads {
		threadMap[t.ThreadID] = t
	}
	sorted := make([]threadmodel.Thread, len(ids))
	timestamps := make([]time.Time, len(ids))
	for i, id := range ids {
		sorted[i] = threadMap[id]
		timestamps[i] = orderMap[id]
	}

	ctx := context.Background()
	if err := r.Cache.Set(ctx, key, cachedResult{
		Threads:    sorted,
		Timestamps: timestamps,
		Total:      total,
	}, r.ttl); err == nil {
		r.Cache.RegisterCacheKey(ctx, InteractionUserListTag(r.table, userID), key)
		for _, threadID := range ids {
			if threadID != 0 {
				r.Cache.RegisterCacheKey(ctx, r.InteractionThreadTag(threadID), key)
			}
		}
	}

	return sorted, timestamps, total, nil
}

func InteractionThreadCacheKeysKey(table string, threadID uint) string {
	return fmt.Sprintf("%s:thread:%d:cache_keys", table, threadID)
}
