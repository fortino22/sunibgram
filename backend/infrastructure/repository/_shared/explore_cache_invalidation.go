package infrarepoShared

import (
	"context"

	"github.com/sunibgram/infrastructure/cache"
)

func InvalidateExploreCache(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.InvalidateTag(ctx, ExploreGlobalTag())
}

func InvalidateExplorePostCaches(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.InvalidateTags(ctx, ExploreTopTag(), ExploreLatestTag())
}

func InvalidateExploreTopCache(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.InvalidateTag(ctx, ExploreTopTag())
}
