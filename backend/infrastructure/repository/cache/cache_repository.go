package cacherepo

import (
	"context"
	"time"

	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

type CacheRepository struct {
	cache *cache.Cache
}

func NewCacheRepository(c *cache.Cache) cacherepoiface.ICacheRepository {
	return &CacheRepository{cache: c}
}

func (r *CacheRepository) ClearAll() error {
	return r.cache.FlushDB(context.Background())
}

func (r *CacheRepository) Get(ctx context.Context, key string, dest any) error {
	return r.cache.Get(ctx, key, dest)
}

func (r *CacheRepository) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.cache.Set(ctx, key, value, ttl)
}

func (r *CacheRepository) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	return r.cache.SetNX(ctx, key, value, ttl)
}

func (r *CacheRepository) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	return r.cache.MGet(ctx, keys)
}

func (r *CacheRepository) MSet(ctx context.Context, entries map[string]any, ttl time.Duration) error {
	return r.cache.MSet(ctx, entries, ttl)
}

func (r *CacheRepository) Delete(ctx context.Context, keys ...string) error {
	return r.cache.Delete(ctx, keys...)
}

func (r *CacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	return r.cache.DeleteByPattern(ctx, pattern)
}

func (r *CacheRepository) DeleteByPatterns(ctx context.Context, patterns ...string) error {
	return r.cache.DeleteByPatterns(ctx, patterns...)
}

func (r *CacheRepository) IncrWithWindow(ctx context.Context, key string, window time.Duration) (int64, error) {
	count, _, err := r.cache.IncrWithWindow(ctx, key, window)
	return count, err
}

func (r *CacheRepository) Incr(ctx context.Context, key string) (int64, error) {
	return r.cache.Incr(ctx, key)
}

func (r *CacheRepository) IncrByWithWindow(ctx context.Context, key string, n int64, window time.Duration) (int64, error) {
	count, _, err := r.cache.IncrByWithWindow(ctx, key, n, window)
	return count, err
}

func (r *CacheRepository) GetCounter(ctx context.Context, key string) (int64, error) {
	count, _, err := r.cache.CounterTTL(ctx, key)
	return count, err
}

func (r *CacheRepository) RegisterCacheKey(ctx context.Context, tag, cacheKey string) error {
	return r.cache.RegisterCacheKey(ctx, tag, cacheKey)
}

func (r *CacheRepository) RegisterCacheKeyMany(ctx context.Context, tags []string, cacheKey string) error {
	return r.cache.RegisterCacheKeyMany(ctx, tags, cacheKey)
}

func (r *CacheRepository) InvalidateTag(ctx context.Context, tag string) error {
	return r.cache.InvalidateTag(ctx, tag)
}

func (r *CacheRepository) InvalidateTags(ctx context.Context, tags ...string) error {
	return r.cache.InvalidateTags(ctx, tags...)
}

func (r *CacheRepository) InvalidateGlobalThreadFeeds(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.GlobalThreadFeedsTag())
}

func (r *CacheRepository) InvalidateViewerThreadFeeds(ctx context.Context, viewerID uint) error {
	if viewerID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.ViewerThreadFeedsTag(viewerID))
}

func (r *CacheRepository) InvalidateUserFeeds(ctx context.Context, userID uint) error {
	if userID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.UserFeedsTag(userID))
}

func (r *CacheRepository) InvalidateUserForYouFeeds(ctx context.Context, userID uint) error {
	if userID == 0 {
		return nil
	}
	infrarepoShared.IncrForYouVersion(ctx, r.cache, userID)
	return nil
}

func (r *CacheRepository) InvalidateThreadChain(ctx context.Context, rootID uint) error {
	if rootID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadChainTag(rootID))
}

func (r *CacheRepository) InvalidateThreadCommunity(ctx context.Context, communityID uint) error {
	if communityID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadCommunityTag(communityID))
}

func (r *CacheRepository) InvalidateEventThreadsByEvent(ctx context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.EventThreadsByEventTag(eventID))
}

func (r *CacheRepository) InvalidateEventLists(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.EventListsTag())
}

func (r *CacheRepository) InvalidateInteractionUserList(ctx context.Context, table string, userID uint) error {
	if table == "" || userID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.InteractionUserListTag(table, userID))
}

func (r *CacheRepository) InvalidateCommunityMembersGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityMembersGlobalTag())
}

func (r *CacheRepository) InvalidateCommunityListsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityListsGlobalTag())
}

func (r *CacheRepository) InvalidateCommunityListsByUser(ctx context.Context, userID uint) error {
	if userID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityListsByUserTag(userID))
}

func (r *CacheRepository) InvalidateCommunityMembers(ctx context.Context, communityID uint) error {
	if communityID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityMembersTag(communityID))
}

func (r *CacheRepository) InvalidateCommunityRequests(ctx context.Context, communityID uint) error {
	if communityID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityRequestsTag(communityID))
}

func (r *CacheRepository) InvalidateCommunityIDGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityIDGlobalTag())
}

func (r *CacheRepository) InvalidateThreadRootsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadRootsGlobalTag())
}

func (r *CacheRepository) InvalidateThreadChainGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadChainGlobalTag())
}

func (r *CacheRepository) InvalidateThreadCommunityGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadCommunityGlobalTag())
}

func (r *CacheRepository) InvalidateThreadAllCommunitiesGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadAllCommunitiesGlobalTag())
}

func (r *CacheRepository) InvalidateThreadMyCommunitiesGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadMyCommunitiesGlobalTag())
}

func (r *CacheRepository) InvalidateThreadFollowingGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.ThreadFollowingGlobalTag())
}

func (r *CacheRepository) InvalidateCommunityThreadsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityThreadsGlobalTag())
}

func (r *CacheRepository) InvalidateEventThreadsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.EventThreadsGlobalTag())
}

func (r *CacheRepository) RegisterTrendingHashtag(ctx context.Context, cacheKey string) error {
	return r.cache.RegisterCacheKey(ctx, infrarepoShared.TrendingHashtagsTag(), cacheKey)
}

func (r *CacheRepository) InvalidateTrendingHashtags(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.TrendingHashtagsTag())
}

func (r *CacheRepository) RegisterReportsID(ctx context.Context, cacheKey string) error {
	return r.cache.RegisterCacheKey(ctx, infrarepoShared.ReportsIDTag(), cacheKey)
}

func (r *CacheRepository) InvalidateReportsID(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.ReportsIDTag())
}

func (r *CacheRepository) InvalidateCommunityRequestsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, infrarepoShared.CommunityRequestsGlobalTag())
}
