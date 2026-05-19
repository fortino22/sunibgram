package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

func (r *threadCached) ExistsRepost(userID, originalThreadID uint) (bool, error) {
	var count int64
	err := r.db.readDB.Model(&model.Thread{}).
		Where("user_id = ? AND type = ? AND original_thread_id = ?", userID, model.ThreadTypeRepost, originalThreadID).
		Count(&count).Error
	return count > 0, err
}

func (r *threadCached) AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(threadIDs))
	if len(threadIDs) == 0 {
		return result, nil
	}

	var ids []uint
	err := r.db.readDB.Model(&model.Thread{}).
		Select("original_thread_id").
		Where("user_id = ? AND type = ? AND original_thread_id IN ?", userID, model.ThreadTypeRepost, threadIDs).
		Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

func (r *threadCached) DeleteRepost(userID, originalThreadID uint) error {
	var repost model.Thread
	if err := r.db.readDB.
		Where("user_id = ? AND type = ? AND original_thread_id = ?", userID, model.ThreadTypeRepost, originalThreadID).
		First(&repost).Error; err != nil {
		return err
	}

	if err := r.db.writeDB.Delete(&repost).Error; err != nil {
		return err
	}

	r.db.writeDB.Model(&model.Thread{}).
		Where("thread_id = ?", originalThreadID).
		UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
	r.patchCachedCount(originalThreadID, "repost_count", -1)
	patchEventCounterByThreadID(r.db.writeDB, r.cache, originalThreadID, "repost_count", -1)

	ctx := context.Background()
	r.invalidate(ctx, repost.ThreadID, nil, nil)
	r.invalidateFeedCachesByThreadID(ctx, repost.ThreadID)
	r.invalidateInteractions(ctx, originalThreadID)
	r.invalidateUserFeedCache(ctx, repost.UserID)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, repost.UserID)
	r.invalidateForYouFeedCachesForFollowers(ctx, repost.UserID)
	r.invalidateTrendingFeed(ctx)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	var orig model.Thread
	if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, originalThreadID).Error; err == nil {
		r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
	}
	invalidateExploreTopCache(ctx, r.cache)
	return nil
}
