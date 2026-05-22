package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	likemodel "github.com/sunibgram/domain/model/like"
	threadmodel "github.com/sunibgram/domain/model/thread"
	likerepoiface "github.com/sunibgram/domain/repository/like"
	"github.com/sunibgram/infrastructure/cache"
)

type LikeRepository struct {
	base       *interactionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewLikeRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) likerepoiface.ILikeRepository {
	return &LikeRepository{
		base:       newInteractionRepo(writeDB, readDB, c, "likes", &threadDB{writeDB: writeDB, readDB: readDB}),
		patchCount: patchCount,
	}
}

func (r *LikeRepository) Save(like *likemodel.Like) error {
	if err := r.base.saveRecord(like); err != nil {
		return err
	}
	r.base.writeDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", like.ThreadID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	patchEventCounterByThreadID(r.base.writeDB, r.base.cache, like.ThreadID, "like_count", 1)
	patchHashtagCounterByThreadID(r.base.writeDB, r.base.cache, like.ThreadID, "like_count", 1)
	ctx := context.Background()
	r.base.invalidateUser(ctx, like.UserID)
	r.patchCount(like.ThreadID, "like_count", 1)
	ViewerFlagAdd(r.base.cache, like.UserID, "liked", like.ThreadID)
	_ = r.base.cache.InvalidateTags(ctx,
		communityThreadsGlobalTag(),
		threadRootsGlobalTag(),
	)
	invalidateExploreTopCache(ctx, r.base.cache)
	return nil
}

func (r *LikeRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &likemodel.Like{}); err != nil {
		return err
	}
	r.base.writeDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	patchEventCounterByThreadID(r.base.writeDB, r.base.cache, threadID, "like_count", -1)
	patchHashtagCounterByThreadID(r.base.writeDB, r.base.cache, threadID, "like_count", -1)
	ctx := context.Background()
	r.base.invalidateUser(ctx, userID)
	r.patchCount(threadID, "like_count", -1)
	ViewerFlagRemove(r.base.cache, userID, "liked", threadID)
	_ = r.base.cache.InvalidateTag(ctx, communityThreadsGlobalTag())
	invalidateExploreTopCache(ctx, r.base.cache)
	return nil
}

func (r *LikeRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &likemodel.Like{})
}

func (r *LikeRepository) AreThreadsLiked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return r.base.areThreadsInteracted(userID, threadIDs)
}

func (r *LikeRepository) AllLikedThreadIDs(userID uint) ([]uint, error) {
	return r.base.allUserThreadIDs(userID)
}

func (r *LikeRepository) FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
