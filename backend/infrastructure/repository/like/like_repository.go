package likerepo

import (
	"context"
	"time"

	"gorm.io/gorm"

	likemodel "github.com/sunibgram/domain/model/like"
	threadmodel "github.com/sunibgram/domain/model/thread"
	likerepoiface "github.com/sunibgram/domain/repository/like"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	hashtagrepo "github.com/sunibgram/infrastructure/repository/hashtag"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
)

type LikeRepository struct {
	base       *infrarepoShared.InteractionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewLikeRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) likerepoiface.ILikeRepository {
	return &LikeRepository{
		base:       infrarepoShared.NewInteractionRepo(writeDB, readDB, c, "likes", threadrepo.PreloadThreadDetails),
		patchCount: patchCount,
	}
}

func (r *LikeRepository) Save(like *likemodel.Like) error {
	if err := r.base.SaveRecord(like); err != nil {
		return err
	}
	r.base.WriteDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", like.ThreadID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	infrarepoShared.PatchEventCounterByThreadID(r.base.WriteDB, r.base.Cache, like.ThreadID, "like_count", 1)
	hashtagrepo.PatchHashtagCounterByThreadID(r.base.WriteDB, r.base.Cache, like.ThreadID, "like_count", 1)
	ctx := context.Background()
	r.base.InvalidateUser(ctx, like.UserID)
	r.patchCount(like.ThreadID, "like_count", 1)
	infrarepoShared.ViewerFlagAdd(r.base.Cache, like.UserID, "liked", like.ThreadID)
	_ = r.base.Cache.InvalidateTags(ctx,
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.ThreadRootsGlobalTag(),
	)
	infrarepoShared.InvalidateExploreTopCache(ctx, r.base.Cache)
	return nil
}

func (r *LikeRepository) Delete(userID, threadID uint) error {
	if err := r.base.DeleteRecord(userID, threadID, &likemodel.Like{}); err != nil {
		return err
	}
	r.base.WriteDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	infrarepoShared.PatchEventCounterByThreadID(r.base.WriteDB, r.base.Cache, threadID, "like_count", -1)
	hashtagrepo.PatchHashtagCounterByThreadID(r.base.WriteDB, r.base.Cache, threadID, "like_count", -1)
	ctx := context.Background()
	r.base.InvalidateUser(ctx, userID)
	r.patchCount(threadID, "like_count", -1)
	infrarepoShared.ViewerFlagRemove(r.base.Cache, userID, "liked", threadID)
	_ = r.base.Cache.InvalidateTags(ctx,
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.ThreadRootsGlobalTag(),
	)
	infrarepoShared.InvalidateExploreTopCache(ctx, r.base.Cache)
	return nil
}

func (r *LikeRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.Exists(userID, threadID, &likemodel.Like{})
}

func (r *LikeRepository) AreThreadsLiked(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return r.base.AreThreadsInteracted(userID, threadIDs)
}

func (r *LikeRepository) AllLikedThreadIDs(userID uint) ([]uint, error) {
	return r.base.AllUserThreadIDs(userID)
}

func (r *LikeRepository) FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	return r.base.FindByUserID(userID, page, limit)
}
