package repostrepo

import (
	"context"
	"time"

	"gorm.io/gorm"

	repostmodel "github.com/sunibgram/domain/model/repost"
	threadmodel "github.com/sunibgram/domain/model/thread"
	repostrepoiface "github.com/sunibgram/domain/repository/repost"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	hashtagrepo "github.com/sunibgram/infrastructure/repository/hashtag"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
)

type RepostRepository struct {
	base       *infrarepoShared.InteractionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewRepostRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) repostrepoiface.IRepostRepository {
	return &RepostRepository{
		base:       infrarepoShared.NewInteractionRepo(writeDB, readDB, c, "reposts", threadrepo.PreloadThreadDetails),
		patchCount: patchCount,
	}
}

func (r *RepostRepository) Save(repost *repostmodel.Repost) error {
	if err := r.base.SaveRecord(repost); err != nil {
		return err
	}
	r.base.WriteDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", repost.ThreadID).
		UpdateColumn("repost_count", gorm.Expr("repost_count + 1"))
	infrarepoShared.PatchEventCounterByThreadID(r.base.WriteDB, r.base.Cache, repost.ThreadID, "repost_count", 1)
	hashtagrepo.PatchHashtagCounterByThreadID(r.base.WriteDB, r.base.Cache, repost.ThreadID, "repost_count", 1)
	ctx := context.Background()
	r.base.InvalidateUser(ctx, repost.UserID)
	r.patchCount(repost.ThreadID, "repost_count", 1)
	_ = r.base.Cache.InvalidateTags(ctx,
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.ThreadRootsGlobalTag(),
	)
	infrarepoShared.InvalidateExploreTopCache(ctx, r.base.Cache)
	return nil
}

func (r *RepostRepository) Delete(userID, threadID uint) error {
	if err := r.base.DeleteRecord(userID, threadID, &repostmodel.Repost{}); err != nil {
		return err
	}
	r.base.WriteDB.Model(&threadmodel.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
	infrarepoShared.PatchEventCounterByThreadID(r.base.WriteDB, r.base.Cache, threadID, "repost_count", -1)
	hashtagrepo.PatchHashtagCounterByThreadID(r.base.WriteDB, r.base.Cache, threadID, "repost_count", -1)
	ctx := context.Background()
	r.base.InvalidateUser(ctx, userID)
	r.patchCount(threadID, "repost_count", -1)
	_ = r.base.Cache.InvalidateTag(ctx, infrarepoShared.CommunityThreadsGlobalTag())
	infrarepoShared.InvalidateExploreTopCache(ctx, r.base.Cache)
	return nil
}

func (r *RepostRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.Exists(userID, threadID, &repostmodel.Repost{})
}

func (r *RepostRepository) AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return r.base.AreThreadsInteracted(userID, threadIDs)
}

func (r *RepostRepository) FindByUserID(userID uint, page, limit int) ([]threadmodel.Thread, []time.Time, int64, error) {
	return r.base.FindByUserID(userID, page, limit)
}
