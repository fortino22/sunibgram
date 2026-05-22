package bookmarkrepo

import (
	"context"
	"time"

	"gorm.io/gorm"

	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
	threadmodel "github.com/sunibgram/domain/model/thread"
	bookmarkrepoiface "github.com/sunibgram/domain/repository/bookmark"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
)

type BookmarkRepository struct {
	base       *infrarepoShared.InteractionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewBookmarkRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) bookmarkrepoiface.IBookmarkRepository {
	return &BookmarkRepository{
		base:       infrarepoShared.NewInteractionRepo(writeDB, readDB, c, "bookmarks", threadrepo.PreloadThreadDetails),
		patchCount: patchCount,
	}
}

func (r *BookmarkRepository) Save(bookmark *bookmarkmodel.Bookmark) error {
	if err := r.base.SaveRecord(bookmark); err != nil {
		return err
	}
	r.base.WriteDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", bookmark.ThreadID).
		UpdateColumn("bookmark_count", gorm.Expr("bookmark_count + 1"))
	ctx := context.Background()
	r.base.InvalidateUser(ctx, bookmark.UserID)
	r.patchCount(bookmark.ThreadID, "bookmark_count", 1)
	infrarepoShared.ViewerFlagAdd(r.base.Cache, bookmark.UserID, "bookmarked", bookmark.ThreadID)
	_ = r.base.Cache.InvalidateTags(ctx,
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.ThreadRootsGlobalTag(),
	)
	infrarepoShared.InvalidateExploreCache(ctx, r.base.Cache)
	return nil
}

func (r *BookmarkRepository) Delete(userID, threadID uint) error {
	if err := r.base.DeleteRecord(userID, threadID, &bookmarkmodel.Bookmark{}); err != nil {
		return err
	}
	r.base.WriteDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("bookmark_count", gorm.Expr("GREATEST(bookmark_count - 1, 0)"))
	ctx := context.Background()
	r.base.InvalidateUser(ctx, userID)
	r.patchCount(threadID, "bookmark_count", -1)
	infrarepoShared.ViewerFlagRemove(r.base.Cache, userID, "bookmarked", threadID)
	_ = r.base.Cache.InvalidateTag(ctx, infrarepoShared.CommunityThreadsGlobalTag())
	infrarepoShared.InvalidateExploreCache(ctx, r.base.Cache)
	return nil
}

func (r *BookmarkRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.Exists(userID, threadID, &bookmarkmodel.Bookmark{})
}

func (r *BookmarkRepository) AreThreadsBookmarked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return r.base.AreThreadsInteracted(userID, threadIDs)
}

func (r *BookmarkRepository) AllBookmarkedThreadIDs(userID uint) ([]uint, error) {
	return r.base.AllUserThreadIDs(userID)
}

func (r *BookmarkRepository) FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	return r.base.FindByUserID(userID, page, limit)
}
