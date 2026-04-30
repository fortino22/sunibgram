package repository

import (
	"context"
	"sort"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type threadCached struct {
	db        *threadDB
	cache     *cache.Cache
	ttl       time.Duration
	likes     *interactionRepo
	bookmarks *interactionRepo
	reposts   *interactionRepo
}

type threadDB struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewThreadRepository(writeDB, readDB *gorm.DB, c *cache.Cache) (domainrepo.IThreadRepository, func(uint, string, int)) {
	r := &threadCached{
		db:        &threadDB{writeDB: writeDB, readDB: readDB},
		cache:     c,
		ttl:       cache.DefaultTTL,
		likes:     newInteractionRepo(writeDB, readDB, c, "likes"),
		bookmarks: newInteractionRepo(writeDB, readDB, c, "bookmarks"),
		reposts:   newInteractionRepo(writeDB, readDB, c, "reposts"),
	}
	return r, r.patchCachedCount
}

func (r *threadCached) Save(thread *model.Thread) error {
	if err := r.db.writeDB.Create(thread).Error; err != nil {
		return err
	}
	r.invalidate(context.Background(), thread.ThreadID, thread.RootID)
	return nil
}

func (r *threadCached) FindByID(id uint) (*model.Thread, error) {
	ctx := context.Background()
	var thread model.Thread
	key := threadKey(id)
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := r.db.readDB.Preload("User").First(&thread, id).Error; err != nil {
			return nil, err
		}
		r.cache.Set(ctx, key, thread, r.ttl)
	}
	return &thread, nil
}

func (r *threadCached) FindRoots(page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadRootsKey(page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).Where("parent_id IS NULL").Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User").
			Where("parent_id IS NULL").
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadChainKey(rootID, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).Where("thread_id = ? OR root_id = ?", rootID, rootID).Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User").
			Where("thread_id = ? OR root_id = ?", rootID, rootID).
			Order("created_at ASC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Threads, result.Total, nil
}

type userFeedEntry struct {
	threadID   uint
	sortedAt   time.Time
	repostedAt *time.Time
	parentID   *uint
}

func (r *threadCached) buildUserFeedEntries(userID uint, page, limit int) (pageEntries []userFeedEntry, threadMap map[uint]model.Thread, total int64, err error) {
	var ownedThreads []model.Thread
	if err = r.db.readDB.Select("thread_id, created_at, parent_id").
		Where("user_id = ?", userID).
		Find(&ownedThreads).Error; err != nil {
		return
	}

	entries := make([]userFeedEntry, 0, len(ownedThreads))
	for _, t := range ownedThreads {
		entries = append(entries, userFeedEntry{
			threadID: t.ThreadID,
			sortedAt: t.CreatedAt,
			parentID: t.ParentID,
		})
	}

	var reposts []model.Repost
	if err = r.db.readDB.Where("user_id = ?", userID).Find(&reposts).Error; err != nil {
		return
	}
	for _, rp := range reposts {
		ts := rp.CreatedAt
		entries = append(entries, userFeedEntry{
			threadID:   rp.ThreadID,
			sortedAt:   ts,
			repostedAt: &ts,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].sortedAt.After(entries[j].sortedAt)
	})

	total = int64(len(entries))
	offset := (page - 1) * limit
	if offset >= len(entries) {
		return nil, nil, total, nil
	}
	end := offset + limit
	if end > len(entries) {
		end = len(entries)
	}
	pageEntries = entries[offset:end]

	ids := make([]uint, len(pageEntries))
	for i, e := range pageEntries {
		ids[i] = e.threadID
	}
	var threads []model.Thread
	if err = r.db.readDB.Preload("User").Where("thread_id IN ?", ids).Find(&threads).Error; err != nil {
		return
	}
	threadMap = make(map[uint]model.Thread, len(threads))
	for _, t := range threads {
		threadMap[t.ThreadID] = t
	}
	return
}

func (r *threadCached) FindPostsByUserID(userID uint, page, limit int) ([]model.Thread, []*time.Time, int64, error) {
	pageEntries, threadMap, total, err := r.buildUserFeedEntries(userID, page, limit)
	if err != nil || len(pageEntries) == 0 {
		return nil, nil, total, err
	}

	sorted := make([]model.Thread, 0, len(pageEntries))
	timestamps := make([]*time.Time, 0, len(pageEntries))
	for _, e := range pageEntries {
		if t, ok := threadMap[e.threadID]; ok {
			sorted = append(sorted, t)
			timestamps = append(timestamps, e.repostedAt)
		}
	}
	return sorted, timestamps, total, nil
}

func (r *threadCached) FindReplyTabByUserID(userID uint, page, limit int) ([]model.Thread, []*time.Time, []*model.Thread, int64, error) {
	pageEntries, threadMap, total, err := r.buildUserFeedEntries(userID, page, limit)
	if err != nil || len(pageEntries) == 0 {
		return nil, nil, nil, total, err
	}

	// fetch parents for reply entries on this page
	parentIDSet := make(map[uint]struct{})
	for _, e := range pageEntries {
		if e.parentID != nil {
			parentIDSet[*e.parentID] = struct{}{}
		}
	}
	parentMap := make(map[uint]*model.Thread)
	if len(parentIDSet) > 0 {
		pids := make([]uint, 0, len(parentIDSet))
		for id := range parentIDSet {
			pids = append(pids, id)
		}
		var parentThreads []model.Thread
		if err = r.db.readDB.Preload("User").Where("thread_id IN ?", pids).Find(&parentThreads).Error; err != nil {
			return nil, nil, nil, 0, err
		}
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	sortedThreads := make([]model.Thread, 0, len(pageEntries))
	repostTimestamps := make([]*time.Time, 0, len(pageEntries))
	parents := make([]*model.Thread, 0, len(pageEntries))
	for _, e := range pageEntries {
		t, ok := threadMap[e.threadID]
		if !ok {
			continue
		}
		sortedThreads = append(sortedThreads, t)
		repostTimestamps = append(repostTimestamps, e.repostedAt)
		if e.parentID != nil {
			parents = append(parents, parentMap[*e.parentID])
		} else {
			parents = append(parents, nil)
		}
	}
	return sortedThreads, repostTimestamps, parents, total, nil
}

func (r *threadCached) Update(thread *model.Thread) error {
	if err := r.db.writeDB.Model(thread).
		UpdateColumns(map[string]any{
			"content":    thread.Content,
			"updated_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
		return err
	}
	ctx := context.Background()
	r.cache.Delete(ctx, threadKey(thread.ThreadID))
	r.invalidate(ctx, thread.ThreadID, thread.RootID)
	r.invalidateInteractions(ctx, thread.ThreadID)
	return nil
}

func (r *threadCached) patchCachedCount(threadID uint, column string, delta int) {
	ctx := context.Background()
	key := threadKey(threadID)
	var thread model.Thread
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := r.db.readDB.Preload("User").First(&thread, threadID).Error; err != nil {
			return
		}
	}
	switch column {
	case "like_count":
		thread.LikeCount = max(thread.LikeCount+delta, 0)
	case "repost_count":
		thread.RepostCount = max(thread.RepostCount+delta, 0)
	case "bookmark_count":
		thread.BookmarkCount = max(thread.BookmarkCount+delta, 0)
	}
	r.cache.Set(ctx, key, thread, r.ttl)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	r.cache.DeleteByPattern(ctx, "thread:chain:*")
}

func (r *threadCached) Delete(id uint) error {
	var thread model.Thread
	r.db.readDB.Select("thread_id, root_id").First(&thread, id)

	ctx := context.Background()
	r.invalidate(ctx, id, thread.RootID)
	r.invalidateInteractions(ctx, id)

	r.likes.deleteByThreadID(id)
	r.bookmarks.deleteByThreadID(id)
	r.reposts.deleteByThreadID(id)

	if err := r.db.writeDB.Delete(&model.Thread{}, id).Error; err != nil {
		return err
	}
	return nil
}
