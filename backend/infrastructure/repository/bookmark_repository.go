package repository

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type BookmarkRepository struct {
	base       *interactionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewBookmarkRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) domainrepo.IBookmarkRepository {
	return &BookmarkRepository{
		base:       newInteractionRepo(writeDB, readDB, c, "bookmarks", &threadDB{writeDB: writeDB, readDB: readDB}),
		patchCount: patchCount,
	}
}

func (r *BookmarkRepository) Save(bookmark *model.Bookmark) error {
	if err := r.base.saveRecord(bookmark); err != nil {
		return err
	}
	r.base.writeDB.Model(&model.Thread{}).Where("thread_id = ?", bookmark.ThreadID).
		UpdateColumn("bookmark_count", gorm.Expr("bookmark_count + 1"))
	ctx := context.Background()
	r.base.invalidateUser(ctx, bookmark.UserID)
	r.patchCount(bookmark.ThreadID, "bookmark_count", 1)
	r.base.cache.DeleteByPattern(ctx, "community:threads:top:*")
	r.base.cache.DeleteByPattern(ctx, "community:threads:latest:*")
	r.base.cache.DeleteByPattern(ctx, "thread:roots:*")
	invalidateExploreCache(ctx, r.base.cache)
	return nil
}

func (r *BookmarkRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &model.Bookmark{}); err != nil {
		return err
	}
	r.base.writeDB.Model(&model.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("bookmark_count", gorm.Expr("GREATEST(bookmark_count - 1, 0)"))
	ctx := context.Background()
	r.base.invalidateUser(ctx, userID)
	r.patchCount(threadID, "bookmark_count", -1)
	r.base.cache.DeleteByPattern(ctx, "community:threads:top:*")
	r.base.cache.DeleteByPattern(ctx, "community:threads:latest:*")
	invalidateExploreCache(ctx, r.base.cache)
	return nil
}

func (r *BookmarkRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &model.Bookmark{})
}

func (r *BookmarkRepository) AreThreadsBookmarked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return r.base.areThreadsInteracted(userID, threadIDs)
}

func (r *BookmarkRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
