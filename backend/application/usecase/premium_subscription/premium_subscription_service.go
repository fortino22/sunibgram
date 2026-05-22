package premiumsubscriptionuc

import (
	"context"
	"time"

	premiumsubscriptionmodel "github.com/sunibgram/domain/model/premium_subscription"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	premiumsubscriptionrepoiface "github.com/sunibgram/domain/repository/premium_subscription"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	premiumsubscriptionsvciface "github.com/sunibgram/domain/service/premium_subscription"
)

type PremiumSubscriptionService struct {
	userRepo                userrepoiface.IUserRepository
	premiumSubscriptionRepo premiumsubscriptionrepoiface.IPremiumSubscriptionRepository
	communityRepo           communityrepoiface.ICommunityRepository
	cacheRepo               cacherepoiface.ICacheRepository
}

func NewPremiumSubscriptionService(
	userRepo userrepoiface.IUserRepository,
	premiumSubscriptionRepo premiumsubscriptionrepoiface.IPremiumSubscriptionRepository,
	communityRepo communityrepoiface.ICommunityRepository,
	cacheRepo cacherepoiface.ICacheRepository,
) premiumsubscriptionsvciface.IPremiumSubscriptionService {
	return &PremiumSubscriptionService{
		userRepo:                userRepo,
		premiumSubscriptionRepo: premiumSubscriptionRepo,
		communityRepo:           communityRepo,
		cacheRepo:               cacheRepo,
	}
}

func (s *PremiumSubscriptionService) ActivateTrial(userID uint) (*premiumsubscriptionmodel.PremiumSubscription, error) {
	isTrial, err := s.premiumSubscriptionRepo.IsTrialActivated(userID)
	if err != nil {
		return nil, err
	}
	if isTrial {
		return nil, premiumsubscriptionrepoiface.ErrTrialAlreadyActivated
	}

	now := time.Now()
	subscription := &premiumsubscriptionmodel.PremiumSubscription{
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
	if err := s.communityRepo.DeactivateByOwnerIDs([]uint{userID}); err != nil {
		return err
	}
	if err := s.userRepo.UpdatePremium(userID, false); err != nil {
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
	_ = s.cacheRepo.InvalidateCommunityThreadsGlobal(ctx)
	_ = s.cacheRepo.InvalidateThreadAllCommunitiesGlobal(ctx)
	_ = s.cacheRepo.InvalidateCommunityMembersGlobal(ctx)
	_ = s.cacheRepo.InvalidateCommunityRequestsGlobal(ctx)
}
