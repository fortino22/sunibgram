package repository

import (
	"context"
	"fmt"

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
	r.cache.DeleteByPatterns(ctx,
		"community:threads:top:*",
		"community:threads:latest:*",
		"thread:all-communities:*",
		"thread:my-communities:*",
	)
}

// trackFeedCacheKey registers a cache key under a thread's feed tag so
// invalidateFeedCachesByThreadID can drop every feed that contains the thread
// in one O(set size) sweep — no SCAN. Pure delegation to the generic tag API.
func (r *threadCached) trackFeedCacheKey(ctx context.Context, threadID uint, key string) {
	if threadID == 0 {
		return
	}
	r.cache.RegisterCacheKey(ctx, threadFeedsTag(threadID), key)
}

// trackUserFeedKey registers a cache key under a user's feed tag.
// Use for thread:userfeed:<uid>:*, thread:usermedia:<uid>:*, thread:following:<uid>:* —
// all user-scoped feed cache families. Lets invalidateUserFeedCache drop them
// without SCAN.
func (r *threadCached) trackUserFeedKey(ctx context.Context, userID uint, key string) {
	if userID == 0 {
		return
	}
	r.cache.RegisterCacheKey(ctx, userFeedsTag(userID), key)
}

// trackForYouKey registers a foryou source/newcount cache key under the user's
// foryou tag.
func (r *threadCached) trackForYouKey(ctx context.Context, userID uint, key string) {
	if userID == 0 {
		return
	}
	r.cache.RegisterCacheKey(ctx, userForYouTag(userID), key)
}

func (r *threadCached) invalidateFeedCachesByThreadID(ctx context.Context, threadID uint) {
	_ = r.cache.InvalidateTag(ctx, threadFeedsTag(threadID))
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

// invalidateUserFeedCache drops thread:userfeed/usermedia/following entries
// for the user via reverse-index tag — no SCAN. Coarser blast radius than the
// legacy per-pattern call (also nukes media + following) but each Set site
// re-registers cheaply on miss, so the over-invalidation is harmless.
func (r *threadCached) invalidateUserFeedCache(ctx context.Context, userID uint) {
	_ = r.cache.InvalidateTag(ctx, userFeedsTag(userID))
}

func (r *threadCached) invalidateUserMediaCache(ctx context.Context, userID uint) {
	_ = r.cache.InvalidateTag(ctx, userFeedsTag(userID))
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
	_ = r.cache.InvalidateTag(ctx, userForYouTag(userID))
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
	// User-scoped: reverse-index tag (O(set size), no SCAN).
	_ = r.cache.InvalidateTag(ctx, userFeedsTag(userID))
	// Global feed families: low-cardinality pagination variants, SCAN is fine.
	r.cache.DeleteByPatterns(ctx,
		"thread:following:*",
		"thread:roots:*",
		"thread:my-communities:*",
		"thread:all-communities:*",
		"community:threads:*",
		"event:threads:*",
	)
	r.invalidateTrendingFeed(ctx)
	r.invalidateForYouFeedCacheForUser(ctx, userID)
	r.invalidateForYouFeedCachesForFollowers(ctx, userID)
	invalidateExploreCache(ctx, r.cache)
}
