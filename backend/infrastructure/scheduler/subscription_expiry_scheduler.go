package scheduler

import (
	"context"
	"log"
	"time"

	premiumsubscriptionsvciface "github.com/sunibgram/domain/service/premium_subscription"
)

const subscriptionExpiryInterval = 1 * time.Hour

// RunSubscriptionExpiry expires premium subscriptions every hour. Honors ctx
// cancellation for graceful shutdown.
func RunSubscriptionExpiry(ctx context.Context, svc premiumsubscriptionsvciface.IPremiumSubscriptionService) {
	ticker := time.NewTicker(subscriptionExpiryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := svc.ExpireSubscriptions(); err != nil {
				log.Printf("ExpireSubscriptions error: %v", err)
			}
		}
	}
}
