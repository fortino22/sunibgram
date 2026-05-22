package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
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

func threadForYouFeedPattern(userID uint) string {
	return fmt.Sprintf("feed:foryou:user:%d:*", userID)
}

func communityThreadsQueryKey(sort, query string, userID uint, page, limit int) string {
	return fmt.Sprintf("community:threads:%s:%s:user:%d:page:%d:limit:%d", sort, query, userID, page, limit)
}

func (r *threadCached) invalidateCommunityThreadQueryCaches(ctx context.Context) {
	r.cache.DeleteByPattern(ctx, "community:threads:top:*")
	r.cache.DeleteByPattern(ctx, "community:threads:latest:*")
	r.cache.DeleteByPattern(ctx, "thread:all-communities:*")
	r.cache.DeleteByPattern(ctx, "thread:my-communities:*")
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
	r.cache.Delete(ctx, threadKey(threadID), parentThreadKey(threadID))

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

// invalidateForYouFeedCacheForUser drops the per-user candidate caches and the
// "new posts" counter cache so the next feed read fetches fresh content from
// the DB. The session seed is intentionally preserved here so a user scrolling
// the feed does not see the page order shuffle when a like/comment fires.
// Seed invalidation only happens via InvalidateForYouSeed, called from the
// explicit refresh endpoint.
func (r *threadCached) invalidateForYouFeedCacheForUser(ctx context.Context, userID uint) {
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("feed:foryou:user:%d:src:*", userID))
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("feed:foryou:user:%d:newcount:*", userID))
}

func (r *threadCached) invalidateForYouFeedCachesForFollowers(ctx context.Context, userID uint) {
	var followerIDs []uint
	if err := r.db.readDB.WithContext(ctx).
		Model(&model.UserRelationship{}).
		Where("following_id = ?", userID).
		Pluck("follower_id", &followerIDs).Error; err != nil {
		return
	}
	for _, followerID := range followerIDs {
		r.invalidateForYouFeedCacheForUser(ctx, followerID)
	}
}

func (r *threadCached) invalidateForYouFeedCachesForCommunity(ctx context.Context, communityID uint) {
	var memberIDs []uint
	if err := r.db.readDB.WithContext(ctx).
		Model(&model.CommunityMember{}).
		Where("community_id = ?", communityID).
		Pluck("user_id", &memberIDs).Error; err != nil {
		return
	}
	for _, memberID := range memberIDs {
		r.invalidateForYouFeedCacheForUser(ctx, memberID)
	}
}

func (r *threadCached) invalidateTrendingFeed(ctx context.Context) {
	r.cache.Delete(ctx, "feed:trending:global")
}

func (r *threadCached) InvalidateUserFeedCaches(ctx context.Context, userID uint) {
	r.invalidateUserFeedCache(ctx, userID)
	r.invalidateUserMediaCache(ctx, userID)
	r.invalidateTrendingFeed(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, userID)
	r.invalidateForYouFeedCachesForFollowers(ctx, userID)
	r.invalidateFollowingFeedCache(ctx)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	r.cache.DeleteByPattern(ctx, "thread:my-communities:*")
	r.cache.DeleteByPattern(ctx, "thread:all-communities:*")
	r.cache.DeleteByPattern(ctx, "community:threads:*")
	r.cache.DeleteByPattern(ctx, "event:threads:*")
	invalidateExploreCache(ctx, r.cache)
}
