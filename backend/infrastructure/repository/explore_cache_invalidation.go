package repository

import (
	"context"

	"github.com/sunibgram/infrastructure/cache"
)

func invalidateExploreCache(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.InvalidateTag(ctx, exploreGlobalTag())
}

func invalidateExplorePostCaches(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.InvalidateTags(ctx, exploreTopTag(), exploreLatestTag())
}

func invalidateExploreTopCache(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.InvalidateTag(ctx, exploreTopTag())
}
