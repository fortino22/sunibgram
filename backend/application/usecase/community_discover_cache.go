package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sunibgram/domain/repository"
)

const discoverCommunitiesCacheKey = "discover:communities"

var discoverCommunitiesTTL = 5 * time.Minute

func communitySearchCacheKey(communityName string) string {
	return fmt.Sprintf("communities:search:%s", strings.ToLower(strings.TrimSpace(communityName)))
}

func invalidateDiscoverCommunitiesCache(ctx context.Context, cacheRepo repository.ICacheRepository) {
	if cacheRepo == nil {
		return
	}
	_ = cacheRepo.Delete(ctx, discoverCommunitiesCacheKey)
	_ = cacheRepo.DeleteByPattern(ctx, "communities:search:*")
	_ = cacheRepo.DeleteByPattern(ctx, "community:id:*")
}
