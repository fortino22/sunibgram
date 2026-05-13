package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type premiumSubscriptionRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewPremiumSubscriptionRepository(writeDB, readDB *gorm.DB) domainrepo.IPremiumSubscriptionRepository {
	return &premiumSubscriptionRepository{writeDB: writeDB, readDB: readDB}
}

func (r *premiumSubscriptionRepository) Create(subscription *model.PremiumSubscription) error {
	return r.writeDB.Create(subscription).Error
}

func (r *premiumSubscriptionRepository) Upsert(subscription *model.PremiumSubscription) error {
	return r.writeDB.Model(&model.PremiumSubscription{}).
		Where("user_id = ?", subscription.UserID).
		Assign(model.PremiumSubscription{
			IsTrial:               subscription.IsTrial,
			SubscriptionStartDate: subscription.SubscriptionStartDate,
			SubscriptionEndDate:   subscription.SubscriptionEndDate,
		}).
		FirstOrCreate(subscription).Error
}

func (r *premiumSubscriptionRepository) IsTrialActivated(userID uint) (bool, error) {
	var count int64
	if err := r.readDB.Model(&model.PremiumSubscription{}).
		Where("user_id = ? AND is_trial = ?", userID, true).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *premiumSubscriptionRepository) FindExpiredOwnerIDs() ([]uint, error) {
	var userIDs []uint
	err := r.readDB.Model(&model.PremiumSubscription{}).
		Joins("JOIN users ON users.user_id = premium_subscriptions.user_id").
		Where("premium_subscriptions.subscription_end_date < ? AND users.is_premium = ?", time.Now(), true).
		Distinct("premium_subscriptions.user_id").
		Pluck("premium_subscriptions.user_id", &userIDs).Error
	return userIDs, err
}

func (r *premiumSubscriptionRepository) ExpireForDev(userID uint) error {
	return r.writeDB.Model(&model.PremiumSubscription{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"subscription_end_date": gorm.Expr("NOW() - INTERVAL '1 second'"),
			"is_trial":              false,
		}).Error
}
