package usecase

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type PremiumSubscriptionService struct {
	userRepo                repository.IUserRepository
	premiumSubscriptionRepo repository.IPremiumSubscriptionRepository
	communityRepo           repository.ICommunityRepository
	cacheRepo               repository.ICacheRepository
}

func NewPremiumSubscriptionService(
	userRepo repository.IUserRepository,
	premiumSubscriptionRepo repository.IPremiumSubscriptionRepository,
	communityRepo repository.ICommunityRepository,
	cacheRepo repository.ICacheRepository,
) service.IPremiumSubscriptionService {
	return &PremiumSubscriptionService{
		userRepo:                userRepo,
		premiumSubscriptionRepo: premiumSubscriptionRepo,
		communityRepo:           communityRepo,
		cacheRepo:               cacheRepo,
	}
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
	if err := s.premiumSubscriptionRepo.Upsert(subscription); err != nil {
		return nil, err
	}
	if err := s.communityRepo.ReactivateByOwnerID(userID); err != nil {
		return nil, err
	}
	s.invalidateCommunityThreadQueryCaches()

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
	s.invalidateCommunityThreadQueryCaches()
	return nil
}

func (s *PremiumSubscriptionService) ExpireForDev(userID uint) error {
	if err := s.premiumSubscriptionRepo.ExpireForDev(userID); err != nil {
		return err
	}
	if err := s.ExpireSubscriptions(); err != nil {
		return err
	}
	s.invalidateCommunityThreadQueryCaches()
	return nil
}

func (s *PremiumSubscriptionService) invalidateCommunityThreadQueryCaches() {
	if s.cacheRepo == nil {
		return
	}
	ctx := context.Background()
	_ = s.cacheRepo.DeleteByPattern(ctx, "community:threads:top:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "community:threads:latest:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "thread:all-communities:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "community:requests:*")
	_ = s.cacheRepo.DeleteByPattern(ctx, "community:users:*")
}
