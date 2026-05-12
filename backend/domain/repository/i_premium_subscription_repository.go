package repository

import (
	"errors"

	"github.com/sunibgram/domain/model"
)

var ErrTrialAlreadyActivated = errors.New("trial subscription already activated")

type IPremiumSubscriptionRepository interface {
	Create(subscription *model.PremiumSubscription) error
	IsTrialActivated(userID uint) (bool, error)
}
