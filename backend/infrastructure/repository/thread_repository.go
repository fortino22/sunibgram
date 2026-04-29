package repository

import (
	"context"
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
