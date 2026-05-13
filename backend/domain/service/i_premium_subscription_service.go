package service

import "github.com/sunibgram/domain/model"

type IPremiumSubscriptionService interface {
	ActivateTrial(userID uint) (*model.PremiumSubscription, error)
	ExpireSubscriptions() error
	ExpireForDev(userID uint) error
}
