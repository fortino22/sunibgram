package repository

import (
	"context"
	"fmt"

	"github.com/sunibgram/infrastructure/cache"
)

func notificationSenderKey(userID uint) string {
	return fmt.Sprintf("notification:sender:%d", userID)
}

func invalidateNotificationSenderCache(ctx context.Context, c *cache.Cache, userID uint) {
	if c == nil {
		return
	}
	_ = c.Delete(ctx, notificationSenderKey(userID))
}

func invalidateNotificationSenderCachePattern(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	_ = c.DeleteByPattern(ctx, "notification:sender:*")
}
