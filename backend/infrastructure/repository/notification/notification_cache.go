package notificationrepo

import (
	"context"
	"fmt"

	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

func notificationSenderKey(userID uint) string {
	return fmt.Sprintf("notification:sender:%d", userID)
}

func InvalidateNotificationSenderCache(ctx context.Context, c *cache.Cache, userID uint) {
	if c == nil {
		return
	}
	_ = c.Delete(ctx, notificationSenderKey(userID))
}

func InvalidateNotificationSenderCachePattern(ctx context.Context, c *cache.Cache) {
	if c == nil {
		return
	}
	// infrarepoShared.NotificationSenderGlobalTag is registered at every SetSenderCache call
	// site — drops every cached sender entry in O(set size).
	_ = c.InvalidateTag(ctx, infrarepoShared.NotificationSenderGlobalTag())
}
