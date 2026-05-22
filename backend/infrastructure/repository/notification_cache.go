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
	// notificationSenderGlobalTag is registered at every SetSenderCache call
	// site — drops every cached sender entry in O(set size).
	_ = c.InvalidateTag(ctx, notificationSenderGlobalTag())
}
