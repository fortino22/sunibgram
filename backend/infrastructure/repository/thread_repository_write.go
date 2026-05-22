package repository

import (
	"context"
	"errors"

	"github.com/sunibgram/domain/model"
	"gorm.io/gorm"
)

func (r *threadCached) Save(thread *model.Thread) error {
	if err := r.validateThreadEventExists(thread.EventID); err != nil {
		return err
	}
	if err := r.db.writeDB.Create(thread).Error; err != nil {
		return err
	}
	if err := r.syncThreadEvent(thread.ThreadID, thread.EventID, thread.EventID != nil); err != nil {
		return err
	}
	ctx := context.Background()
	var repostInheritedCommunityID *uint
	if thread.Type == model.ThreadTypeRepost && thread.OriginalThreadID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.OriginalThreadID).
			UpdateColumn("repost_count", gorm.Expr("repost_count + 1"))
		r.patchCachedCount(*thread.OriginalThreadID, "repost_count", 1)
		patchEventCounterByThreadID(r.db.writeDB, r.cache, *thread.OriginalThreadID, "repost_count", 1)
		r.invalidateInteractions(ctx, *thread.OriginalThreadID)
		ViewerFlagAdd(r.cache, thread.UserID, "reposted", *thread.OriginalThreadID)
		var orig model.Thread
		if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, *thread.OriginalThreadID).Error; err == nil {
			r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
			repostInheritedCommunityID = orig.CommunityID
		}
	}
	if thread.ParentID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.ParentID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
		r.patchCachedCount(*thread.ParentID, "comment_count", 1)
	}
	r.invalidate(ctx, thread.ThreadID, thread.RootID, thread.CommunityID)
	r.invalidateUserFeedCache(ctx, thread.UserID)
	r.invalidateUserMediaCache(ctx, thread.UserID)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, thread.UserID)
	r.invalidateForYouFeedCachesForFollowers(ctx, thread.UserID)
	effectiveCommunityID := thread.CommunityID
	if effectiveCommunityID == nil {
		effectiveCommunityID = repostInheritedCommunityID
	}
	if effectiveCommunityID != nil {
		r.invalidateForYouFeedCachesForCommunity(ctx, *effectiveCommunityID)
		r.invalidateCommunityThreadQueryCaches(ctx)
	}
	r.invalidateTrendingFeed(ctx)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	invalidateExplorePostCaches(ctx, r.cache)
	return nil
}

func (r *threadCached) InvalidateThread(threadID uint) {
	ctx := context.Background()
	thread, err := r.FindByID(threadID)
	if err != nil {
		r.cache.Delete(ctx, threadKey(threadID), parentThreadKey(threadID))
		r.invalidateTrendingFeed(ctx)
		return
	}
	r.invalidate(ctx, threadID, thread.RootID, thread.CommunityID)
	r.invalidateFeedCachesByThreadID(ctx, threadID)
	r.invalidateCommunityThreadQueryCaches(ctx)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	r.invalidateInteractions(ctx, threadID)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
	invalidateExplorePostCaches(ctx, r.cache)
}

func (r *threadCached) Update(thread *model.Thread) error {
	if thread.EventIDUpdated {
		if err := r.validateThreadEventExists(thread.EventID); err != nil {
			return err
		}
	}
	if err := r.db.writeDB.Model(thread).
		UpdateColumns(map[string]any{
			"content":    thread.Content,
			"updated_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
		return err
	}
	if err := r.syncThreadEvent(thread.ThreadID, thread.EventID, thread.EventIDUpdated); err != nil {
		return err
	}
	ctx := context.Background()
	r.cache.Delete(ctx, threadKey(thread.ThreadID), parentThreadKey(thread.ThreadID))
	r.invalidate(ctx, thread.ThreadID, thread.RootID, thread.CommunityID)
	r.invalidateFeedCachesByThreadID(ctx, thread.ThreadID)
	r.invalidateInteractions(ctx, thread.ThreadID)
	r.invalidateUserFeedCache(ctx, thread.UserID)
	r.invalidateUserMediaCache(ctx, thread.UserID)
	r.invalidateCommunityThreadQueryCaches(ctx)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	invalidateExplorePostCaches(ctx, r.cache)
	return nil
}

func (r *threadCached) UpdateReplyPermission(threadID uint, replyPermission model.ReplyPermission) error {
	if err := r.db.writeDB.Model(&model.Thread{}).
		Where("thread_id = ?", threadID).
		UpdateColumns(map[string]any{
			"reply_permission": replyPermission,
			"updated_at":       gorm.Expr("NOW()"),
		}).Error; err != nil {
		return err
	}
	ctx := context.Background()
	r.cache.Delete(ctx, threadKey(threadID), parentThreadKey(threadID))
	return nil
}

func (r *threadCached) validateThreadEventExists(eventID *string) error {
	if eventID == nil || *eventID == "" {
		return nil
	}
	var count int64
	if err := r.db.readDB.Model(&model.Event{}).Where("id = ?", *eventID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("event not found")
	}
	return nil
}

func (r *threadCached) syncThreadEvent(threadID uint, eventID *string, replace bool) error {
	if !replace {
		return nil
	}
	ctx := context.Background()
	oldEventID, hadOldEvent := eventIDForThread(r.db.readDB, threadID)
	var stats model.Thread
	_ = r.db.readDB.Select("like_count", "repost_count").First(&stats, threadID).Error
	if err := r.db.writeDB.Where("thread_id = ?", threadID).Delete(&model.ThreadEvent{}).Error; err != nil {
		return err
	}
	if eventID == nil || *eventID == "" {
		if hadOldEvent {
			patchEventCounters(r.db.writeDB, r.cache, oldEventID, -1, -stats.LikeCount, -stats.RepostCount)
			refreshEventParticipantCount(r.db.writeDB, r.cache, oldEventID)
		}
		r.cache.DeleteByPattern(ctx, "event:threads:*")
		return nil
	}
	if err := r.db.writeDB.Create(&model.ThreadEvent{ThreadID: threadID, EventID: *eventID}).Error; err != nil {
		return err
	}
	if !hadOldEvent || oldEventID != *eventID {
		if hadOldEvent {
			patchEventCounters(r.db.writeDB, r.cache, oldEventID, -1, -stats.LikeCount, -stats.RepostCount)
			refreshEventParticipantCount(r.db.writeDB, r.cache, oldEventID)
		}
		patchEventCounters(r.db.writeDB, r.cache, *eventID, 1, stats.LikeCount, stats.RepostCount)
		refreshEventParticipantCount(r.db.writeDB, r.cache, *eventID)
	}
	r.cache.DeleteByPattern(ctx, "event:threads:*")
	return nil
}

func (r *threadCached) patchCachedCount(threadID uint, column string, delta int) {
	ctx := context.Background()
	key := threadKey(threadID)
	var thread model.Thread
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := preloadThreadDetails(r.db.readDB).First(&thread, threadID).Error; err != nil {
			return
		}
	}
	switch column {
	case "like_count":
		thread.LikeCount = max(thread.LikeCount+delta, 0)
	case "repost_count":
		thread.RepostCount = max(thread.RepostCount+delta, 0)
	case "bookmark_count":
		thread.BookmarkCount = max(thread.BookmarkCount+delta, 0)
	case "comment_count":
		thread.CommentCount = max(thread.CommentCount+delta, 0)
	}
	r.cache.Set(ctx, key, thread, r.ttl)
	r.invalidateFeedCachesByThreadID(ctx, threadID)
	r.invalidateForYouFeedCacheForUser(ctx, thread.UserID)
	r.invalidateForYouFeedCachesForFollowers(ctx, thread.UserID)
	if thread.CommunityID != nil {
		r.invalidateForYouFeedCachesForCommunity(ctx, *thread.CommunityID)
	}
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
	r.cache.Delete(ctx, "feed:trending:global")
}

func (r *threadCached) Delete(id uint) error {
	var thread model.Thread
	r.db.readDB.Select("thread_id, user_id, parent_id, root_id, type, original_thread_id, community_id, like_count, repost_count").First(&thread, id)

	ctx := context.Background()
	if thread.Type == model.ThreadTypeRepost && thread.OriginalThreadID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.OriginalThreadID).
			UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
		r.patchCachedCount(*thread.OriginalThreadID, "repost_count", -1)
		patchEventCounterByThreadID(r.db.writeDB, r.cache, *thread.OriginalThreadID, "repost_count", -1)
		r.invalidateInteractions(ctx, *thread.OriginalThreadID)
		var orig model.Thread
		if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, *thread.OriginalThreadID).Error; err == nil {
			r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
		}
	}
	if thread.ParentID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.ParentID).
			UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)"))
		r.patchCachedCount(*thread.ParentID, "comment_count", -1)
	}

	r.invalidate(ctx, id, thread.RootID, thread.CommunityID)
	r.invalidateInteractions(ctx, id)
	r.invalidateUserFeedCache(ctx, thread.UserID)
	r.invalidateUserMediaCache(ctx, thread.UserID)
	r.invalidateFeedCachesByThreadID(ctx, id)
	r.invalidateCommunityThreadQueryCaches(ctx)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")

	r.likes.deleteByThreadID(id)
	r.bookmarks.deleteByThreadID(id)
	r.reposts.deleteByThreadID(id)
	eventID, hasEvent := eventIDForThread(r.db.readDB, id)
	if hasEvent {
		patchEventCounter(r.db.writeDB, r.cache, eventID, "thread_count", -1)
		patchEventCounter(r.db.writeDB, r.cache, eventID, "like_count", -thread.LikeCount)
		patchEventCounter(r.db.writeDB, r.cache, eventID, "repost_count", -thread.RepostCount)
	}

	if err := r.db.writeDB.Delete(&model.Thread{}, id).Error; err != nil {
		return err
	}
	if hasEvent {
		refreshEventParticipantCount(r.db.writeDB, r.cache, eventID)
	}
	invalidateExplorePostCaches(ctx, r.cache)
	return nil
}
