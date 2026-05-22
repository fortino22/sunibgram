package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
	threadmodel "github.com/sunibgram/domain/model/thread"
	bookmarkrepoiface "github.com/sunibgram/domain/repository/bookmark"
	"github.com/sunibgram/infrastructure/cache"
)

type BookmarkRepository struct {
	base       *interactionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewBookmarkRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) bookmarkrepoiface.IBookmarkRepository {
	return &BookmarkRepository{
		base:       newInteractionRepo(writeDB, readDB, c, "bookmarks", &threadDB{writeDB: writeDB, readDB: readDB}),
		patchCount: patchCount,
	}
}

func (r *BookmarkRepository) Save(bookmark *bookmarkmodel.Bookmark) error {
	if err := r.base.saveRecord(bookmark); err != nil {
		return err
	}
	r.base.writeDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", bookmark.ThreadID).
		UpdateColumn("bookmark_count", gorm.Expr("bookmark_count + 1"))
	ctx := context.Background()
	r.base.invalidateUser(ctx, bookmark.UserID)
	r.patchCount(bookmark.ThreadID, "bookmark_count", 1)
	ViewerFlagAdd(r.base.cache, bookmark.UserID, "bookmarked", bookmark.ThreadID)
	_ = r.base.cache.InvalidateTags(ctx,
		communityThreadsGlobalTag(),
		threadRootsGlobalTag(),
	)
	invalidateExploreCache(ctx, r.base.cache)
	return nil
}

func (r *BookmarkRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &bookmarkmodel.Bookmark{}); err != nil {
		return err
	}
	r.base.writeDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("bookmark_count", gorm.Expr("GREATEST(bookmark_count - 1, 0)"))
	ctx := context.Background()
	r.base.invalidateUser(ctx, userID)
	r.patchCount(threadID, "bookmark_count", -1)
	ViewerFlagRemove(r.base.cache, userID, "bookmarked", threadID)
	_ = r.base.cache.InvalidateTag(ctx, communityThreadsGlobalTag())
	invalidateExploreCache(ctx, r.base.cache)
	return nil
}

func (r *BookmarkRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &bookmarkmodel.Bookmark{})
}

func (r *BookmarkRepository) AreThreadsBookmarked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return r.base.areThreadsInteracted(userID, threadIDs)
}

func (r *BookmarkRepository) AllBookmarkedThreadIDs(userID uint) ([]uint, error) {
	return r.base.allUserThreadIDs(userID)
}

func (r *BookmarkRepository) FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
