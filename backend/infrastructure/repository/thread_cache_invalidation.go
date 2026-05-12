package repository

import (
	"context"
	"fmt"
	"time"
)

func threadKey(id uint) string {
	return fmt.Sprintf("thread:id:%d", id)
}

func threadRootsKey(page, limit int) string {
	return fmt.Sprintf("thread:roots:page:%d:limit:%d", page, limit)
}

func threadChainKey(rootID uint, page, limit int) string {
	return fmt.Sprintf("thread:chain:%d:page:%d:limit:%d", rootID, page, limit)
}

func threadCommunityKey(communityID uint, page, limit int, sort string) string {
	return fmt.Sprintf("thread:community:%d:sort:%s:page:%d:limit:%d", communityID, sort, page, limit)
}

func threadUserFeedKey(userID uint, tab string, requesterID uint, page, limit int) string {
	return fmt.Sprintf("thread:userfeed:%d:%s:page:%d:limit:%d:req:%d", userID, tab, page, limit, requesterID)
}

func threadMediaFeedKey(userID, requesterID uint, page, limit int) string {
	return fmt.Sprintf("thread:usermedia:%d:page:%d:limit:%d:req:%d", userID, page, limit, requesterID)
}

func threadFollowingFeedKey(userID uint, page, limit int) string {
	return fmt.Sprintf("thread:following:%d:page:%d:limit:%d", userID, page, limit)
}

func threadFeedKeysKey(threadID uint) string {
	return fmt.Sprintf("thread:feed_keys:%d", threadID)
}

func (r *threadCached) trackFeedCacheKey(ctx context.Context, threadID uint, key string) {
	if threadID == 0 {
		return
	}
	r.cache.SAdd(ctx, threadFeedKeysKey(threadID), key, 24*time.Hour)
}

func (r *threadCached) invalidateFeedCachesByThreadID(ctx context.Context, threadID uint) {
	setKey := threadFeedKeysKey(threadID)
	if cacheKeys, err := r.cache.SMembers(ctx, setKey); err == nil && len(cacheKeys) > 0 {
		r.cache.Delete(ctx, cacheKeys...)
	}
	r.cache.Delete(ctx, setKey)
}

func (r *threadCached) invalidate(ctx context.Context, threadID uint, rootID *uint, communityID *uint) {
	r.cache.Delete(ctx, threadKey(threadID))

	treeID := threadID
	if rootID != nil && *rootID != 0 {
		treeID = *rootID
	}
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:chain:%d:*", treeID))

	if communityID != nil && *communityID != 0 {
		r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:community:%d:*", *communityID))
	}
}

func (r *threadCached) invalidateInteractions(ctx context.Context, threadID uint) {
	r.likes.invalidateByThreadID(ctx, threadID)
	r.bookmarks.invalidateByThreadID(ctx, threadID)
	r.reposts.invalidateByThreadID(ctx, threadID)
}

func (r *threadCached) invalidateUserFeedCache(ctx context.Context, userID uint) {
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:userfeed:%d:*", userID))
}

func (r *threadCached) invalidateUserMediaCache(ctx context.Context, userID uint) {
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:usermedia:%d:*", userID))
}

func (r *threadCached) invalidateFollowingFeedCache(ctx context.Context) {
	r.cache.DeleteByPattern(ctx, "thread:following:*")
}
