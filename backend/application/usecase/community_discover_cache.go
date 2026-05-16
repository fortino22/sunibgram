package usecase

import (
	"context"
	"time"

	"github.com/sunibgram/domain/repository"
)

const discoverCommunitiesCacheKey = "discover:communities"

var discoverCommunitiesTTL = 5 * time.Minute

func invalidateDiscoverCommunitiesCache(ctx context.Context, cacheRepo repository.ICacheRepository) {
	if cacheRepo == nil {
		return
	}
	_ = cacheRepo.Delete(ctx, discoverCommunitiesCacheKey)
}
