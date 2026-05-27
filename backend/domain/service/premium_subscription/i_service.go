package premiumsubscriptionsvciface

import premiumsubscriptionmodel "github.com/sunibgram/domain/model/premium_subscription"

type IPremiumSubscriptionService interface {
	ActivateTrial(userID uint) (*premiumsubscriptionmodel.PremiumSubscription, error)
	GetStatus(userID uint) (sub *premiumsubscriptionmodel.PremiumSubscription, isPremium bool, err error)
	ExpireSubscriptions() error
	ExpireForDev(userID uint) error
}
