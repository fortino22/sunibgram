package repository

import (
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

func (r *premiumSubscriptionRepository) IsTrialActivated(userID uint) (bool, error) {
	var count int64
	if err := r.readDB.Model(&model.PremiumSubscription{}).
		Where("user_id = ? AND is_trial = ?", userID, true).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
