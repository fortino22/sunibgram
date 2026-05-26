package threadrepo

import (
	"context"

	"gorm.io/gorm"

	threadmodel "github.com/sunibgram/domain/model/thread"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

func (r *threadCached) ExistsRepost(userID, originalThreadID uint) (bool, error) {
	var count int64
	err := r.db.readDB.Model(&threadmodel.Thread{}).
		Where("user_id = ? AND type = ? AND original_thread_id = ?", userID, threadmodel.ThreadTypeRepost, originalThreadID).
		Count(&count).Error
	return count > 0, err
}

func (r *threadCached) AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(threadIDs))
	if len(threadIDs) == 0 {
		return result, nil
	}

	var ids []uint
	err := r.db.readDB.Model(&threadmodel.Thread{}).
		Select("original_thread_id").
		Where("user_id = ? AND type = ? AND original_thread_id IN ?", userID, threadmodel.ThreadTypeRepost, threadIDs).
		Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

// RepostWrapperIDsByOriginal returns the thread_ids of all REPOST rows whose
// original_thread_id matches. Used to invalidate repost wrapper caches when the
// original thread state changes (like/unlike/bookmark counter on original).
func (r *threadCached) RepostWrapperIDsByOriginal(originalThreadID uint) ([]uint, error) {
	var ids []uint
	err := r.db.readDB.Model(&threadmodel.Thread{}).
		Select("thread_id").
		Where("type = ? AND original_thread_id = ?", threadmodel.ThreadTypeRepost, originalThreadID).
		Scan(&ids).Error
	return ids, err
}

func (r *threadCached) AllRepostedOriginalIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.readDB.Model(&threadmodel.Thread{}).
		Select("original_thread_id").
		Where("user_id = ? AND type = ? AND original_thread_id IS NOT NULL", userID, threadmodel.ThreadTypeRepost).
		Scan(&ids).Error
	return ids, err
}

func (r *threadCached) DeleteRepost(userID, originalThreadID uint) error {
	var repost threadmodel.Thread
	if err := r.db.readDB.
		Where("user_id = ? AND type = ? AND original_thread_id = ?", userID, threadmodel.ThreadTypeRepost, originalThreadID).
		First(&repost).Error; err != nil {
		return err
	}

	if err := r.db.writeDB.Delete(&repost).Error; err != nil {
		return err
	}

	r.db.writeDB.Model(&threadmodel.Thread{}).
		Where("thread_id = ?", originalThreadID).
		UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
	r.patchCachedCount(originalThreadID, "repost_count", -1)
	infrarepoShared.PatchEventCounterByThreadID(r.db.writeDB, r.cache, originalThreadID, "repost_count", -1)

	ctx := context.Background()
	infrarepoShared.ViewerFlagRemove(r.cache, userID, "reposted", originalThreadID)
	r.invalidate(ctx, repost.ThreadID, nil, nil)
	r.invalidateFeedCachesByThreadID(ctx, repost.ThreadID)
	r.invalidateInteractions(ctx, originalThreadID)
	r.invalidateUserFeedCache(ctx, repost.UserID)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, repost.UserID)
	r.invalidateForYouFeedCachesForFollowers(ctx, repost.UserID)
	r.invalidateTrendingFeed(ctx)
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())
	var orig threadmodel.Thread
	if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, originalThreadID).Error; err == nil {
		r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
	}
	infrarepoShared.InvalidateExploreTopCache(ctx, r.cache)
	return nil
}
