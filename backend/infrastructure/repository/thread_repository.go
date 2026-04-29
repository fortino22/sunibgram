package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

// threadCached wraps threadDB with a Redis cache layer.
// Reads go through cache; writes invalidate only the affected keys.
type threadCached struct {
	db    *threadDB
	cache *cache.Cache
	ttl   time.Duration
}

type threadDB struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewThreadRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IThreadRepository {
	return &threadCached{
		db:    &threadDB{writeDB: writeDB, readDB: readDB},
		cache: c,
		ttl:   cache.DefaultTTL,
	}
}

// ── key helpers ──────────────────────────────────────────────────────────────

func threadKey(id uint) string {
	return fmt.Sprintf("thread:id:%d", id)
}

func threadRootsKey(page, limit int) string {
	return fmt.Sprintf("thread:roots:page:%d:limit:%d", page, limit)
}

func threadChainKey(rootID uint, page, limit int) string {
	return fmt.Sprintf("thread:chain:%d:page:%d:limit:%d", rootID, page, limit)
}


func (r *threadCached) invalidate(ctx context.Context, threadID uint, rootID *uint) {
	r.cache.Delete(ctx, threadKey(threadID))
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	if rootID != nil {
		r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:chain:%d:*", *rootID))
	} else {
		r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:chain:%d:*", threadID))
	}
}


func (r *threadCached) Save(thread *model.Thread) error {
	if err := r.db.writeDB.Create(thread).Error; err != nil {
		return err
	}
	r.invalidate(context.Background(), thread.ThreadID, thread.RootID)
	return nil
}

func (r *threadCached) FindByID(id uint) (*model.Thread, error) {
	var thread model.Thread
	key := threadKey(id)
	if err := r.cache.Get(context.Background(), key, &thread); err == nil {
		return &thread, nil
	}

	if err := r.db.readDB.Preload("User").First(&thread, id).Error; err != nil {
		return nil, err
	}
	r.cache.Set(context.Background(), key, thread, r.ttl)
	return &thread, nil
}

func (r *threadCached) FindRoots(page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	var result cached
	key := threadRootsKey(page, limit)
	if err := r.cache.Get(context.Background(), key, &result); err == nil {
		return result.Threads, result.Total, nil
	}

	var threads []model.Thread
	var total int64
	r.db.readDB.Model(&model.Thread{}).Where("parent_id IS NULL").Count(&total)
	offset := (page - 1) * limit
	err := r.db.readDB.Preload("User").
		Where("parent_id IS NULL").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error
	if err != nil {
		return nil, 0, err
	}

	r.cache.Set(context.Background(), key, cached{Threads: threads, Total: total}, r.ttl)
	return threads, total, nil
}

func (r *threadCached) FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	var result cached
	key := threadChainKey(rootID, page, limit)
	if err := r.cache.Get(context.Background(), key, &result); err == nil {
		return result.Threads, result.Total, nil
	}

	var threads []model.Thread
	var total int64
	r.db.readDB.Model(&model.Thread{}).Where("thread_id = ? OR root_id = ?", rootID, rootID).Count(&total)
	offset := (page - 1) * limit
	err := r.db.readDB.Preload("User").
		Where("thread_id = ? OR root_id = ?", rootID, rootID).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Find(&threads).Error
	if err != nil {
		return nil, 0, err
	}

	r.cache.Set(context.Background(), key, cached{Threads: threads, Total: total}, r.ttl)
	return threads, total, nil
}

func (r *threadCached) Update(thread *model.Thread) error {
	if err := r.db.writeDB.Save(thread).Error; err != nil {
		return err
	}

	ctx := context.Background()

	r.cache.Set(ctx, threadKey(thread.ThreadID), *thread, r.ttl)

	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	if thread.RootID != nil {
		r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:chain:%d:*", *thread.RootID))
	} else {
		r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:chain:%d:*", thread.ThreadID))
	}

	return nil
}

func (r *threadCached) Delete(id uint) error {
	var thread model.Thread
	r.db.readDB.Select("thread_id, root_id").First(&thread, id)

	if err := r.db.writeDB.Delete(&model.Thread{}, id).Error; err != nil {
		return err
	}
	r.invalidate(context.Background(), id, thread.RootID)
	return nil
}
