package threadmediarepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	threadmediarepoiface "github.com/sunibgram/domain/repository/thread_media"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

// Thread media is immutable for the life of a thread (set at creation, dropped
// only on delete), so it is cached per-thread and invalidated on mutation.
const threadMediaTTL = 24 * time.Hour

func threadMediaKey(threadID uint) string {
	return fmt.Sprintf("thread:media:%d:list", threadID)
}

type ThreadMediaRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewThreadMediaRepository(writeDB, readDB *gorm.DB, c ...*cache.Cache) threadmediarepoiface.IThreadMediaRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &ThreadMediaRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

func (r *ThreadMediaRepository) SaveAll(media []threadmediamodel.ThreadMedia) error {
	if err := r.writeDB.Create(&media).Error; err != nil {
		return err
	}
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	r.invalidateMedia(uniqueThreadIDs(media))
	return nil
}

func (r *ThreadMediaRepository) FindByThreadID(threadID uint) ([]threadmediamodel.ThreadMedia, error) {
	ctx := context.Background()
	if r.cache != nil {
		var cached []threadmediamodel.ThreadMedia
		if err := r.cache.Get(ctx, threadMediaKey(threadID), &cached); err == nil {
			return cached, nil
		}
	}
	var media []threadmediamodel.ThreadMedia
	if err := r.readDB.Where("thread_id = ?", threadID).Find(&media).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		// Cache even an empty slice so threads without media aren't re-queried.
		_ = r.cache.Set(ctx, threadMediaKey(threadID), media, threadMediaTTL)
	}
	return media, nil
}

func (r *ThreadMediaRepository) FindByThreadIDs(threadIDs []uint) ([]threadmediamodel.ThreadMedia, error) {
	var media []threadmediamodel.ThreadMedia
	if len(threadIDs) == 0 {
		return media, nil
	}
	ctx := context.Background()
	missing := threadIDs
	if r.cache != nil {
		keys := make([]string, len(threadIDs))
		for i, id := range threadIDs {
			keys[i] = threadMediaKey(id)
		}
		hits, _ := r.cache.MGet(ctx, keys)
		missing = missing[:0]
		for _, id := range threadIDs {
			if raw, ok := hits[threadMediaKey(id)]; ok {
				var cached []threadmediamodel.ThreadMedia
				if json.Unmarshal(raw, &cached) == nil {
					media = append(media, cached...)
					continue
				}
			}
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return media, nil
	}

	var fetched []threadmediamodel.ThreadMedia
	if err := r.readDB.Where("thread_id IN ?", missing).Find(&fetched).Error; err != nil {
		return nil, err
	}
	media = append(media, fetched...)
	if r.cache != nil {
		// Group by thread; cache every missing id (empty slice if no media) so
		// none is re-queried next time.
		byThread := make(map[uint][]threadmediamodel.ThreadMedia, len(missing))
		for _, id := range missing {
			byThread[id] = []threadmediamodel.ThreadMedia{}
		}
		for _, m := range fetched {
			byThread[m.ThreadID] = append(byThread[m.ThreadID], m)
		}
		entries := make(map[string]any, len(byThread))
		for id, list := range byThread {
			entries[threadMediaKey(id)] = list
		}
		_ = r.cache.MSet(ctx, entries, threadMediaTTL)
	}
	return media, nil
}

func (r *ThreadMediaRepository) DeleteByThreadID(threadID uint) error {
	if err := r.writeDB.Where("thread_id = ?", threadID).Delete(&threadmediamodel.ThreadMedia{}).Error; err != nil {
		return err
	}
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	r.invalidateMedia([]uint{threadID})
	return nil
}

func (r *ThreadMediaRepository) DeleteByIDs(ids []uint) ([]threadmediamodel.ThreadMedia, error) {
	var media []threadmediamodel.ThreadMedia
	if err := r.readDB.Where("media_id IN ?", ids).Find(&media).Error; err != nil {
		return nil, err
	}
	if err := r.writeDB.Where("media_id IN ?", ids).Delete(&threadmediamodel.ThreadMedia{}).Error; err != nil {
		return nil, err
	}
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	r.invalidateMedia(uniqueThreadIDs(media))
	return media, nil
}

// invalidateMedia drops the per-thread media cache for the given thread IDs.
func (r *ThreadMediaRepository) invalidateMedia(threadIDs []uint) {
	if r.cache == nil || len(threadIDs) == 0 {
		return
	}
	keys := make([]string, len(threadIDs))
	for i, id := range threadIDs {
		keys[i] = threadMediaKey(id)
	}
	_ = r.cache.Delete(context.Background(), keys...)
}

func uniqueThreadIDs(media []threadmediamodel.ThreadMedia) []uint {
	seen := make(map[uint]struct{}, len(media))
	ids := make([]uint, 0, len(media))
	for _, m := range media {
		if _, ok := seen[m.ThreadID]; ok {
			continue
		}
		seen[m.ThreadID] = struct{}{}
		ids = append(ids, m.ThreadID)
	}
	return ids
}
