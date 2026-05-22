package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	cacherepoiface "github.com/sunibgram/domain/repository/cache"
)

const discoverCommunitiesCacheKey = "discover:communities"

var discoverCommunitiesTTL = 5 * time.Minute

func communitySearchCacheKey(communityName string) string {
	return fmt.Sprintf("communities:search:%s", strings.ToLower(strings.TrimSpace(communityName)))
}

func invalidateDiscoverCommunitiesCache(ctx context.Context, cacheRepo cacherepoiface.ICacheRepository) {
	if cacheRepo == nil {
		return
	}
	_ = cacheRepo.Delete(ctx, discoverCommunitiesCacheKey)
	// communities:search:* — kept as SCAN; per-unique-term keys cached by
	// search queries are hot-write/cold-invalidate, where SCAN cost beats
	// adding a SADD per write.
	_ = cacheRepo.DeleteByPattern(ctx, "communities:search:*")
	// community:id:v2:* — tag-based; communityIDGlobalTag registered at
	// every Set site in community_repository.go.
	_ = cacheRepo.InvalidateCommunityIDGlobal(ctx)
}
