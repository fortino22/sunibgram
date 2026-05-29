package threadrepo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	threadmodel "github.com/sunibgram/domain/model/thread"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	hashtagrepo "github.com/sunibgram/infrastructure/repository/hashtag"
)

func (r *threadCached) Save(thread *threadmodel.Thread) error {
	if err := r.validateThreadEventExists(thread.EventID); err != nil {
		return err
	}
	if err := r.db.writeDB.Create(thread).Error; err != nil {
		return err
	}
	if err := r.syncThreadEvent(thread.ThreadID, thread.EventID, thread.IsOfficial, thread.EventID != nil); err != nil {
		return err
	}
	ctx := context.Background()
	var repostInheritedCommunityID *uint
	if thread.Type == threadmodel.ThreadTypeRepost && thread.OriginalThreadID != nil {
		r.db.writeDB.Model(&threadmodel.Thread{}).
			Where("thread_id = ?", *thread.OriginalThreadID).
			UpdateColumn("repost_count", gorm.Expr("repost_count + 1"))
		r.patchCachedCount(*thread.OriginalThreadID, "repost_count", 1)
		infrarepoShared.PatchEventCounterByThreadID(r.db.writeDB, r.cache, *thread.OriginalThreadID, "repost_count", 1)
		r.invalidateInteractions(ctx, *thread.OriginalThreadID)
		infrarepoShared.ViewerFlagAdd(r.cache, thread.UserID, "reposted", *thread.OriginalThreadID)
		var orig threadmodel.Thread
		if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, *thread.OriginalThreadID).Error; err == nil {
			r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
			repostInheritedCommunityID = orig.CommunityID
		}
	}
	if thread.ParentID != nil {
		r.db.writeDB.Model(&threadmodel.Thread{}).
			Where("thread_id = ?", *thread.ParentID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
		r.patchCachedCount(*thread.ParentID, "comment_count", 1)
	}
	r.invalidate(ctx, thread.ThreadID, thread.RootID, thread.CommunityID)
	r.invalidateUserFeedCache(ctx, thread.UserID)
	r.invalidateUserMediaCache(ctx, thread.UserID)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, thread.UserID)
	go func(uid uint) {
		defer func() { _ = recover() }()
		r.invalidateForYouFeedCachesForFollowers(context.Background(), uid)
	}(thread.UserID)
	effectiveCommunityID := thread.CommunityID
	if effectiveCommunityID == nil {
		effectiveCommunityID = repostInheritedCommunityID
	}
	if effectiveCommunityID != nil {
		go func(cid uint) {
			defer func() { _ = recover() }()
			r.invalidateForYouFeedCachesForCommunity(context.Background(), cid)
		}(*effectiveCommunityID)
		r.invalidateCommunityThreadQueryCaches(ctx)
	}
	r.invalidateTrendingFeed(ctx)
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())
	infrarepoShared.InvalidateExplorePostCaches(ctx, r.cache)
	return nil
}

func (r *threadCached) InvalidateThread(threadID uint) {
	ctx := context.Background()
	thread, err := r.FindByID(threadID)
	if err != nil {
		r.cache.Delete(ctx, ThreadKey(threadID), ParentThreadKey(threadID))
		r.invalidateTrendingFeed(ctx)
		return
	}
	r.invalidate(ctx, threadID, thread.RootID, thread.CommunityID)
	r.invalidateFeedCachesByThreadID(ctx, threadID)
	r.invalidateCommunityThreadQueryCaches(ctx)
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())
	r.invalidateInteractions(ctx, threadID)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
	infrarepoShared.InvalidateExplorePostCaches(ctx, r.cache)
}

func (r *threadCached) Update(thread *threadmodel.Thread) error {
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
	if err := r.syncThreadEvent(thread.ThreadID, thread.EventID, thread.IsOfficial, thread.EventIDUpdated); err != nil {
		return err
	}
	ctx := context.Background()
	r.cache.Delete(ctx, ThreadKey(thread.ThreadID), ParentThreadKey(thread.ThreadID))
	r.invalidate(ctx, thread.ThreadID, thread.RootID, thread.CommunityID)
	r.invalidateFeedCachesByThreadID(ctx, thread.ThreadID)
	r.invalidateInteractions(ctx, thread.ThreadID)
	r.invalidateUserFeedCache(ctx, thread.UserID)
	r.invalidateUserMediaCache(ctx, thread.UserID)
	r.invalidateCommunityThreadQueryCaches(ctx)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, thread.UserID)
	go func(uid uint) {
		defer func() { _ = recover() }()
		r.invalidateForYouFeedCachesForFollowers(context.Background(), uid)
	}(thread.UserID)
	if thread.CommunityID != nil {
		go func(cid uint) {
			defer func() { _ = recover() }()
			r.invalidateForYouFeedCachesForCommunity(context.Background(), cid)
		}(*thread.CommunityID)
	}
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())
	infrarepoShared.InvalidateExplorePostCaches(ctx, r.cache)
	return nil
}

func (r *threadCached) UpdateReplyPermission(threadID uint, replyPermission threadmodel.ReplyPermission) error {
	if err := r.db.writeDB.Model(&threadmodel.Thread{}).
		Where("thread_id = ?", threadID).
		UpdateColumns(map[string]any{
			"reply_permission": replyPermission,
			"updated_at":       gorm.Expr("NOW()"),
		}).Error; err != nil {
		return err
	}
	ctx := context.Background()
	var thread threadmodel.Thread
	if err := r.db.readDB.
		Select("thread_id, user_id, root_id, community_id").
		First(&thread, threadID).Error; err != nil {
		r.cache.Delete(ctx, ThreadKey(threadID), ParentThreadKey(threadID))
		return nil
	}
	r.invalidate(ctx, thread.ThreadID, thread.RootID, thread.CommunityID)
	r.invalidateFeedCachesByThreadID(ctx, thread.ThreadID)
	r.invalidateInteractions(ctx, thread.ThreadID)
	r.invalidateUserFeedCache(ctx, thread.UserID)
	r.invalidateUserMediaCache(ctx, thread.UserID)
	r.invalidateCommunityThreadQueryCaches(ctx)
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, thread.UserID)
	go func(uid uint) {
		defer func() { _ = recover() }()
		r.invalidateForYouFeedCachesForFollowers(context.Background(), uid)
	}(thread.UserID)
	if thread.CommunityID != nil {
		go func(cid uint) {
			defer func() { _ = recover() }()
			r.invalidateForYouFeedCachesForCommunity(context.Background(), cid)
		}(*thread.CommunityID)
	}
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.EventThreadsGlobalTag())
	infrarepoShared.InvalidateExplorePostCaches(ctx, r.cache)
	return nil
}

func (r *threadCached) validateThreadEventExists(eventID *string) error {
	if eventID == nil || *eventID == "" {
		return nil
	}
	var exists bool
	if err := r.db.readDB.Model(&threadmodel.Event{}).Select("1").Where("id = ?", *eventID).Limit(1).Scan(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return errors.New("event not found")
	}
	return nil
}

func (r *threadCached) syncThreadEvent(threadID uint, eventID *string, isOfficial, replace bool) error {
	if !replace {
		return nil
	}
	ctx := context.Background()
	oldEventID, hadOldEvent := infrarepoShared.EventIDForThread(r.db.readDB, threadID)
	var stats threadmodel.Thread
	_ = r.db.readDB.Select("like_count", "repost_count", "share_count").First(&stats, threadID).Error
	if err := r.db.writeDB.Where("thread_id = ?", threadID).Delete(&threadmodel.ThreadEvent{}).Error; err != nil {
		return err
	}
	if eventID == nil || *eventID == "" {
		if hadOldEvent {
			infrarepoShared.PatchEventCounters(r.db.writeDB, r.cache, oldEventID, -1, -stats.LikeCount, -stats.RepostCount, -stats.ShareCount)
			infrarepoShared.RefreshEventParticipantCount(r.db.writeDB, r.cache, oldEventID)
		}
		_ = r.cache.InvalidateTag(ctx, infrarepoShared.EventThreadsGlobalTag())
		return nil
	}
	if err := r.db.writeDB.Create(&threadmodel.ThreadEvent{ThreadID: threadID, EventID: *eventID, IsOfficial: isOfficial}).Error; err != nil {
		return err
	}
	if !hadOldEvent || oldEventID != *eventID {
		if hadOldEvent {
			infrarepoShared.PatchEventCounters(r.db.writeDB, r.cache, oldEventID, -1, -stats.LikeCount, -stats.RepostCount, -stats.ShareCount)
			infrarepoShared.RefreshEventParticipantCount(r.db.writeDB, r.cache, oldEventID)
		}
		infrarepoShared.PatchEventCounters(r.db.writeDB, r.cache, *eventID, 1, stats.LikeCount, stats.RepostCount, stats.ShareCount)
		infrarepoShared.RefreshEventParticipantCount(r.db.writeDB, r.cache, *eventID)
	}
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.EventThreadsGlobalTag())
	return nil
}

func (r *threadCached) patchCachedCount(threadID uint, column string, delta int) {
	ctx := context.Background()
	key := ThreadKey(threadID)
	var thread threadmodel.Thread
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := PreloadThreadDetails(r.db.readDB).Where("threads.thread_id = ?", threadID).First(&thread).Error; err != nil {
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
	case "share_count":
		thread.ShareCount = max(thread.ShareCount+delta, 0)
	}
	r.cache.Set(ctx, key, thread, r.ttl)
	r.invalidateFeedCachesByThreadID(ctx, threadID)
	r.invalidateForYouFeedCacheForUser(ctx, thread.UserID)
	go func(uid uint) {
		defer func() { _ = recover() }()
		r.invalidateForYouFeedCachesForFollowers(context.Background(), uid)
	}(thread.UserID)
	if thread.CommunityID != nil {
		go func(cid uint) {
			defer func() { _ = recover() }()
			r.invalidateForYouFeedCachesForCommunity(context.Background(), cid)
		}(*thread.CommunityID)
	}
	r.invalidateFollowingFeedCache(ctx)
	r.invalidateTrendingFeed(ctx)
}

func (r *threadCached) IncrementShareCount(threadID uint) error {
	if err := r.db.writeDB.Model(&threadmodel.Thread{}).
		Where("thread_id = ?", threadID).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error; err != nil {
		return err
	}
	infrarepoShared.PatchEventCounterByThreadID(r.db.writeDB, r.cache, threadID, "share_count", 1)
	hashtagrepo.PatchHashtagCounterByThreadID(r.db.writeDB, r.cache, threadID, "share_count", 1)
	r.patchCachedCount(threadID, "share_count", 1)
	ctx := context.Background()
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())
	infrarepoShared.InvalidateExplorePostCaches(ctx, r.cache)
	return nil
}

func (r *threadCached) Delete(id uint) error {
	var thread threadmodel.Thread
	r.db.readDB.Select("thread_id, user_id, parent_id, root_id, type, original_thread_id, community_id, like_count, repost_count").First(&thread, id)

	ctx := context.Background()
	if thread.Type == threadmodel.ThreadTypeRepost && thread.OriginalThreadID != nil {
		r.db.writeDB.Model(&threadmodel.Thread{}).
			Where("thread_id = ?", *thread.OriginalThreadID).
			UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
		r.patchCachedCount(*thread.OriginalThreadID, "repost_count", -1)
		infrarepoShared.PatchEventCounterByThreadID(r.db.writeDB, r.cache, *thread.OriginalThreadID, "repost_count", -1)
		r.invalidateInteractions(ctx, *thread.OriginalThreadID)
		var orig threadmodel.Thread
		if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, *thread.OriginalThreadID).Error; err == nil {
			r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
		}
	}
	if thread.ParentID != nil {
		r.db.writeDB.Model(&threadmodel.Thread{}).
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
	r.invalidateForYouFeedCacheForUser(ctx, thread.UserID)
	go func(uid uint) {
		defer func() { _ = recover() }()
		r.invalidateForYouFeedCachesForFollowers(context.Background(), uid)
	}(thread.UserID)
	if thread.CommunityID != nil {
		go func(cid uint) {
			defer func() { _ = recover() }()
			r.invalidateForYouFeedCachesForCommunity(context.Background(), cid)
		}(*thread.CommunityID)
	}
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())

	r.likes.DeleteByThreadID(id)
	r.bookmarks.DeleteByThreadID(id)
	r.reposts.DeleteByThreadID(id)
	eventID, hasEvent := infrarepoShared.EventIDForThread(r.db.readDB, id)
	if hasEvent {
		infrarepoShared.PatchEventCounter(r.db.writeDB, r.cache, eventID, "thread_count", -1)
		infrarepoShared.PatchEventCounter(r.db.writeDB, r.cache, eventID, "like_count", -thread.LikeCount)
		infrarepoShared.PatchEventCounter(r.db.writeDB, r.cache, eventID, "repost_count", -thread.RepostCount)
	}

	if err := r.db.writeDB.Delete(&threadmodel.Thread{}, id).Error; err != nil {
		return err
	}
	if hasEvent {
		infrarepoShared.RefreshEventParticipantCount(r.db.writeDB, r.cache, eventID)
	}
	infrarepoShared.InvalidateExplorePostCaches(ctx, r.cache)
	return nil
}
