package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
)

const threadFeedCountTTL = 5 * time.Minute

// feedIDsCache stores only thread IDs (and aligned parent IDs) for a paginated
// feed query. Threads are rehydrated via hydrateThreads, which prefers the
// per-thread cache (threadKey) and falls back to a batched DB load on miss.
//
// ParentIDs is aligned 1:1 with IDs. A zero entry means the corresponding
// thread has no parent. ParentIDs is nil for feeds that do not surface
// parents.
type feedIDsCache struct {
	IDs       []uint
	ParentIDs []uint
}

func parentThreadKey(id uint) string {
	return fmt.Sprintf("thread:parent:%d", id)
}

func feedCountKey(key string) string {
	return key + ":count"
}

func (r *threadCached) getCachedCount(ctx context.Context, key string) (int64, bool) {
	var total int64
	if err := r.cache.Get(ctx, feedCountKey(key), &total); err == nil {
		return total, true
	}
	return 0, false
}

func (r *threadCached) setCachedCount(ctx context.Context, key string, total int64) {
	_ = r.cache.Set(ctx, feedCountKey(key), total, threadFeedCountTTL)
}

// hydrateThreads returns threads in the order of ids, sourcing from the
// per-thread cache first and batching DB lookups for misses. Missing rows are
// silently dropped (matches FindByID semantics for evicted entries).
func (r *threadCached) hydrateThreads(ctx context.Context, ids []uint) ([]model.Thread, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	found := make(map[uint]model.Thread, len(ids))
	missing := make([]uint, 0)
	for _, id := range ids {
		if id == 0 {
			continue
		}
		var t model.Thread
		if err := r.cache.Get(ctx, threadKey(id), &t); err == nil {
			found[id] = t
			continue
		}
		missing = append(missing, id)
	}
	if len(missing) > 0 {
		var rows []model.Thread
		if err := preloadThreadDetails(r.db.readDB).
			Where("thread_id IN ?", missing).
			Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			found[rows[i].ThreadID] = rows[i]
			_ = r.cache.Set(ctx, threadKey(rows[i].ThreadID), rows[i], r.ttl)
		}
	}
	out := make([]model.Thread, 0, len(ids))
	for _, id := range ids {
		if t, ok := found[id]; ok {
			out = append(out, t)
		}
	}
	return out, nil
}

// hydrateParents returns a parent slice aligned 1:1 with parentIDs. Entries
// where parentIDs[i] == 0 stay nil. Sources from a separate per-parent cache
// (parentThreadKey) since parents are loaded with the lighter
// preloadThreadParentDetails set.
func (r *threadCached) hydrateParents(ctx context.Context, parentIDs []uint) ([]*model.Thread, error) {
	if len(parentIDs) == 0 {
		return nil, nil
	}
	uniq := make(map[uint]struct{})
	for _, id := range parentIDs {
		if id != 0 {
			uniq[id] = struct{}{}
		}
	}
	found := make(map[uint]*model.Thread, len(uniq))
	missing := make([]uint, 0, len(uniq))
	for id := range uniq {
		var t model.Thread
		if err := r.cache.Get(ctx, parentThreadKey(id), &t); err == nil {
			tc := t
			found[id] = &tc
			continue
		}
		missing = append(missing, id)
	}
	if len(missing) > 0 {
		var rows []model.Thread
		if err := preloadThreadParentDetails(r.db.readDB).
			Where("thread_id IN ?", missing).
			Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			tc := rows[i]
			found[rows[i].ThreadID] = &tc
			_ = r.cache.Set(ctx, parentThreadKey(rows[i].ThreadID), rows[i], r.ttl)
		}
	}
	out := make([]*model.Thread, len(parentIDs))
	for i, id := range parentIDs {
		if id == 0 {
			continue
		}
		out[i] = found[id]
	}
	return out, nil
}

// storeFeedIDs caches the ordered IDs and primes the per-thread / per-parent
// caches so subsequent hits avoid the heavy preload chain. Pass parents=nil
// for feeds that do not surface parents.
func (r *threadCached) storeFeedIDs(
	ctx context.Context,
	key string,
	threads []model.Thread,
	parents []*model.Thread,
	ttl time.Duration,
) {
	entry := feedIDsCache{IDs: make([]uint, len(threads))}
	if parents != nil {
		entry.ParentIDs = make([]uint, len(threads))
	}
	for i, t := range threads {
		entry.IDs[i] = t.ThreadID
		_ = r.cache.Set(ctx, threadKey(t.ThreadID), t, r.ttl)
		if parents != nil && parents[i] != nil {
			entry.ParentIDs[i] = parents[i].ThreadID
			_ = r.cache.Set(ctx, parentThreadKey(parents[i].ThreadID), *parents[i], r.ttl)
		}
	}
	if err := r.cache.Set(ctx, key, entry, ttl); err == nil {
		for _, id := range entry.IDs {
			r.trackFeedCacheKey(ctx, id, key)
		}
	}
}

func (r *threadCached) loadFeedIDs(ctx context.Context, key string) (feedIDsCache, bool) {
	var entry feedIDsCache
	if err := r.cache.Get(ctx, key, &entry); err != nil {
		return entry, false
	}
	return entry, true
}

// hydrateChainThreads is the parent-detail variant of hydrateThreads. Used by
// FindByRootID where rows are loaded with preloadThreadParentDetails (lighter
// preload set). Stores into a distinct cache namespace (parentThreadKey) so
// reads do not conflict with the full-detail per-thread cache.
func (r *threadCached) hydrateChainThreads(ctx context.Context, ids []uint) ([]model.Thread, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	found := make(map[uint]model.Thread, len(ids))
	missing := make([]uint, 0)
	for _, id := range ids {
		if id == 0 {
			continue
		}
		var t model.Thread
		if err := r.cache.Get(ctx, parentThreadKey(id), &t); err == nil {
			found[id] = t
			continue
		}
		missing = append(missing, id)
	}
	if len(missing) > 0 {
		var rows []model.Thread
		if err := preloadThreadParentDetails(r.db.readDB).
			Where("thread_id IN ?", missing).
			Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			found[rows[i].ThreadID] = rows[i]
			_ = r.cache.Set(ctx, parentThreadKey(rows[i].ThreadID), rows[i], r.ttl)
		}
	}
	out := make([]model.Thread, 0, len(ids))
	for _, id := range ids {
		if t, ok := found[id]; ok {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *threadCached) storeChainFeedIDs(ctx context.Context, key string, threads []model.Thread) {
	entry := feedIDsCache{IDs: make([]uint, len(threads))}
	for i, t := range threads {
		entry.IDs[i] = t.ThreadID
		_ = r.cache.Set(ctx, parentThreadKey(t.ThreadID), t, r.ttl)
	}
	if err := r.cache.Set(ctx, key, entry, threadFeedTTL); err == nil {
		for _, id := range entry.IDs {
			r.trackFeedCacheKey(ctx, id, key)
		}
	}
}
