package usecase

import (
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type PremiumSubscriptionService struct {
	userRepo                repository.IUserRepository
	premiumSubscriptionRepo repository.IPremiumSubscriptionRepository
}

func NewPremiumSubscriptionService(userRepo repository.IUserRepository, premiumSubscriptionRepo repository.IPremiumSubscriptionRepository) service.IPremiumSubscriptionService {
	return &PremiumSubscriptionService{userRepo: userRepo, premiumSubscriptionRepo: premiumSubscriptionRepo}
}

func (s *PremiumSubscriptionService) ActivateTrial(userID uint) (*model.PremiumSubscription, error) {
	isTrial, err := s.premiumSubscriptionRepo.IsTrialActivated(userID)
	if err != nil {
		return nil, err
	}
	if isTrial {
		return nil, repository.ErrTrialAlreadyActivated
	}

	now := time.Now()
	subscription := &model.PremiumSubscription{
		UserID:                userID,
		IsTrial:               true,
		SubscriptionStartDate: now,
		SubscriptionEndDate:   now.AddDate(0, 1, 0),
	}

	if err := s.userRepo.UpdatePremium(userID, true); err != nil {
		return nil, err
	}
	if err := s.premiumSubscriptionRepo.Create(subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}
