package repository

import (
	"context"

	"github.com/sunibgram/infrastructure/cache"
)

const exploreCachePattern = "explore:*"

func invalidateExploreCache(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.DeleteByPattern(ctx, exploreCachePattern)
}
