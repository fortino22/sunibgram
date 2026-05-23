package premiumsubscriptionrepoiface

import (
	"errors"

	premiumsubscriptionmodel "github.com/sunibgram/domain/model/premium_subscription"
)

var ErrTrialAlreadyActivated = errors.New("trial subscription already activated")

type IPremiumSubscriptionRepository interface {
	Create(subscription *premiumsubscriptionmodel.PremiumSubscription) error
	Upsert(subscription *premiumsubscriptionmodel.PremiumSubscription) error
	IsTrialActivated(userID uint) (bool, error)
	FindExpiredOwnerIDs() ([]uint, error)
	ExpireForDev(userID uint) error
}
