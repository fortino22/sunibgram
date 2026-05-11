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

	ctx := context.Background()
	r.invalidate(ctx, repost.ThreadID, nil, nil)
	r.invalidateFeedCachesByThreadID(ctx, repost.ThreadID)
	r.invalidateInteractions(ctx, originalThreadID)
	r.invalidateUserFeedCache(ctx, repost.UserID)
	r.invalidateFollowingFeedCache(ctx)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	var orig model.Thread
	if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, originalThreadID).Error; err == nil {
		r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
	}
	invalidateExploreCache(ctx, r.cache)
	return nil
}
