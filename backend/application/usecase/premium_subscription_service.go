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
	communityRepo           repository.ICommunityRepository
}

func NewPremiumSubscriptionService(userRepo repository.IUserRepository, premiumSubscriptionRepo repository.IPremiumSubscriptionRepository, communityRepo repository.ICommunityRepository) service.IPremiumSubscriptionService {
	return &PremiumSubscriptionService{userRepo: userRepo, premiumSubscriptionRepo: premiumSubscriptionRepo, communityRepo: communityRepo}
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

func (s *PremiumSubscriptionService) ExpireSubscriptions() error {
	ownerIDs, err := s.premiumSubscriptionRepo.FindExpiredOwnerIDs()
	if err != nil {
		return err
	}
	if len(ownerIDs) == 0 {
		return nil
	}
	if err := s.communityRepo.DeactivateByOwnerIDs(ownerIDs); err != nil {
		return err
	}
	for _, id := range ownerIDs {
		if err := s.userRepo.UpdatePremium(id, false); err != nil {
			return err
		}
	}
	return nil
}

func (s *PremiumSubscriptionService) ExpireForDev(userID uint) error {
	if err := s.premiumSubscriptionRepo.ExpireForDev(userID); err != nil {
		return err
	}
	return s.ExpireSubscriptions()
}
