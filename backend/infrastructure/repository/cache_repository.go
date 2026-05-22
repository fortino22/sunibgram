package repository

import (
	"context"
	"time"

	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
)

type CacheRepository struct {
	cache *cache.Cache
}

func NewCacheRepository(c *cache.Cache) domainrepo.ICacheRepository {
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
	return r.cache.InvalidateTag(ctx, globalThreadFeedsTag())
}

func (r *CacheRepository) InvalidateViewerThreadFeeds(ctx context.Context, viewerID uint) error {
	if viewerID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, viewerThreadFeedsTag(viewerID))
}

func (r *CacheRepository) InvalidateUserFeeds(ctx context.Context, userID uint) error {
	if userID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, userFeedsTag(userID))
}

func (r *CacheRepository) InvalidateUserForYouFeeds(ctx context.Context, userID uint) error {
	if userID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, userForYouTag(userID))
}

func (r *CacheRepository) InvalidateThreadChain(ctx context.Context, rootID uint) error {
	if rootID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, threadChainTag(rootID))
}

func (r *CacheRepository) InvalidateThreadCommunity(ctx context.Context, communityID uint) error {
	if communityID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, threadCommunityTag(communityID))
}

func (r *CacheRepository) InvalidateEventThreadsByEvent(ctx context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}
	return r.cache.InvalidateTag(ctx, eventThreadsByEventTag(eventID))
}

func (r *CacheRepository) InvalidateEventLists(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, eventListsTag())
}

func (r *CacheRepository) InvalidateInteractionUserList(ctx context.Context, table string, userID uint) error {
	if table == "" || userID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, interactionUserListTag(table, userID))
}

func (r *CacheRepository) InvalidateCommunityMembersGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, communityMembersGlobalTag())
}

func (r *CacheRepository) InvalidateCommunityListsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, communityListsGlobalTag())
}

func (r *CacheRepository) InvalidateCommunityListsByUser(ctx context.Context, userID uint) error {
	if userID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, communityListsByUserTag(userID))
}

func (r *CacheRepository) InvalidateCommunityMembers(ctx context.Context, communityID uint) error {
	if communityID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, communityMembersTag(communityID))
}

func (r *CacheRepository) InvalidateCommunityRequests(ctx context.Context, communityID uint) error {
	if communityID == 0 {
		return nil
	}
	return r.cache.InvalidateTag(ctx, communityRequestsTag(communityID))
}

func (r *CacheRepository) InvalidateCommunityIDGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, communityIDGlobalTag())
}

func (r *CacheRepository) InvalidateThreadRootsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, threadRootsGlobalTag())
}

func (r *CacheRepository) InvalidateThreadChainGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, threadChainGlobalTag())
}

func (r *CacheRepository) InvalidateThreadCommunityGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, threadCommunityGlobalTag())
}

func (r *CacheRepository) InvalidateThreadAllCommunitiesGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, threadAllCommunitiesGlobalTag())
}

func (r *CacheRepository) InvalidateThreadMyCommunitiesGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, threadMyCommunitiesGlobalTag())
}

func (r *CacheRepository) InvalidateThreadFollowingGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, threadFollowingGlobalTag())
}

func (r *CacheRepository) InvalidateCommunityThreadsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, communityThreadsGlobalTag())
}

func (r *CacheRepository) InvalidateEventThreadsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, eventThreadsGlobalTag())
}

func (r *CacheRepository) RegisterTrendingHashtag(ctx context.Context, cacheKey string) error {
	return r.cache.RegisterCacheKey(ctx, trendingHashtagsTag(), cacheKey)
}

func (r *CacheRepository) InvalidateTrendingHashtags(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, trendingHashtagsTag())
}

func (r *CacheRepository) RegisterReportsID(ctx context.Context, cacheKey string) error {
	return r.cache.RegisterCacheKey(ctx, reportsIDTag(), cacheKey)
}

func (r *CacheRepository) InvalidateReportsID(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, reportsIDTag())
}

func (r *CacheRepository) InvalidateCommunityRequestsGlobal(ctx context.Context) error {
	return r.cache.InvalidateTag(ctx, communityRequestsGlobalTag())
}
