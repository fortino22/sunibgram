package repository

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type LikeRepository struct {
	base       *interactionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewLikeRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) domainrepo.ILikeRepository {
	return &LikeRepository{
		base:       newInteractionRepo(writeDB, readDB, c, "likes", &threadDB{writeDB: writeDB, readDB: readDB}),
		patchCount: patchCount,
	}
}

func (r *LikeRepository) Save(like *model.Like) error {
	if err := r.base.saveRecord(like); err != nil {
		return err
	}
	r.base.writeDB.Model(&model.Thread{}).Where("thread_id = ?", like.ThreadID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	patchEventCounterByThreadID(r.base.writeDB, r.base.cache, like.ThreadID, "like_count", 1)
	patchHashtagCounterByThreadID(r.base.writeDB, r.base.cache, like.ThreadID, "like_count", 1)
	ctx := context.Background()
	r.base.invalidateUser(ctx, like.UserID)
	r.patchCount(like.ThreadID, "like_count", 1)
	r.base.cache.DeleteByPattern(ctx, "community:threads:top:*")
	r.base.cache.DeleteByPattern(ctx, "community:threads:latest:*")
	invalidateExploreTopCache(ctx, r.base.cache)
	return nil
}

func (r *LikeRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &model.Like{}); err != nil {
		return err
	}
	r.base.writeDB.Model(&model.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	patchEventCounterByThreadID(r.base.writeDB, r.base.cache, threadID, "like_count", -1)
	patchHashtagCounterByThreadID(r.base.writeDB, r.base.cache, threadID, "like_count", -1)
	ctx := context.Background()
	r.base.invalidateUser(ctx, userID)
	r.patchCount(threadID, "like_count", -1)
	r.base.cache.DeleteByPattern(ctx, "community:threads:top:*")
	r.base.cache.DeleteByPattern(ctx, "community:threads:latest:*")
	invalidateExploreTopCache(ctx, r.base.cache)
	return nil
}

func (r *LikeRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &model.Like{})
}

func (r *LikeRepository) AreThreadsLiked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return r.base.areThreadsInteracted(userID, threadIDs)
}

func (r *LikeRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
