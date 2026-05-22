package repository

import (
	"context"
	"time"
)

type ICacheRepository interface {
	ClearAll() error
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)
	MSet(ctx context.Context, entries map[string]any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	DeleteByPatterns(ctx context.Context, patterns ...string) error

	// Reverse-index tag invalidation (phase 2). RegisterCacheKey associates a
	// cache key with one or more tags on Set. InvalidateTag(s) drops every
	// member key of a tag in O(set size) — no SCAN required. Tag names follow
	// the "tag:<entity>:<id>[:scope]" convention (see infrastructure/repository
	// tag builders). Tag TTL is bounded; longest cache TTL must not exceed it.
	RegisterCacheKey(ctx context.Context, tag, cacheKey string) error
	RegisterCacheKeyMany(ctx context.Context, tags []string, cacheKey string) error
	InvalidateTag(ctx context.Context, tag string) error
	InvalidateTags(ctx context.Context, tags ...string) error

	// High-level tag invalidation surfaces. Hide the tag-name convention
	// behind a stable API so application-layer callers do not need to know
	// the internal tag namespace. Implementations route to InvalidateTag with
	// the matching builder in infrastructure/repository/tag_keys.go.
	InvalidateGlobalThreadFeeds(ctx context.Context) error
	InvalidateViewerThreadFeeds(ctx context.Context, viewerID uint) error
	InvalidateUserFeeds(ctx context.Context, userID uint) error
	InvalidateUserForYouFeeds(ctx context.Context, userID uint) error
	InvalidateThreadChain(ctx context.Context, rootID uint) error
	InvalidateThreadCommunity(ctx context.Context, communityID uint) error
	InvalidateEventThreadsByEvent(ctx context.Context, eventID string) error
	InvalidateEventLists(ctx context.Context) error
	InvalidateInteractionUserList(ctx context.Context, table string, userID uint) error
	InvalidateCommunityMembersGlobal(ctx context.Context) error
	InvalidateCommunityListsGlobal(ctx context.Context) error
	InvalidateCommunityListsByUser(ctx context.Context, userID uint) error
	InvalidateCommunityMembers(ctx context.Context, communityID uint) error
	InvalidateCommunityRequests(ctx context.Context, communityID uint) error
	InvalidateCommunityIDGlobal(ctx context.Context) error

	// Per-family thread-feed globals (precise alternative to
	// InvalidateGlobalThreadFeeds which drops all feed families).
	InvalidateThreadRootsGlobal(ctx context.Context) error
	InvalidateThreadChainGlobal(ctx context.Context) error
	InvalidateThreadCommunityGlobal(ctx context.Context) error
	InvalidateThreadAllCommunitiesGlobal(ctx context.Context) error
	InvalidateThreadMyCommunitiesGlobal(ctx context.Context) error
	InvalidateThreadFollowingGlobal(ctx context.Context) error
	InvalidateCommunityThreadsGlobal(ctx context.Context) error
	InvalidateEventThreadsGlobal(ctx context.Context) error

	// Trending hashtags (read in HashtagService, written by counter
	// patches in hashtag_repository). Tag membership keeps Set + Invalidate
	// symmetric across packages.
	RegisterTrendingHashtag(ctx context.Context, cacheKey string) error
	InvalidateTrendingHashtags(ctx context.Context) error

	// Reports list (admin surfaces) — tag-based wipe of reports:id:* paginated
	// list pages.
	RegisterReportsID(ctx context.Context, cacheKey string) error
	InvalidateReportsID(ctx context.Context) error

	InvalidateCommunityRequestsGlobal(ctx context.Context) error
}
