package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

const defaultExploreLimit = 20

type ExploreService struct {
	exploreRepo  repository.IExploreRepository
	settingsRepo repository.IUserSettingsRepository
}

func NewExploreService(exploreRepo repository.IExploreRepository, settingsRepo repository.IUserSettingsRepository) service.IExploreService {
	return &ExploreService{exploreRepo: exploreRepo, settingsRepo: settingsRepo}
}

func (s *ExploreService) ExploreSearch(ctx context.Context, query string, requesterID *uint) (*model.ExploreResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is required")
	}

	result, err := s.exploreRepo.ExploreSearch(ctx, query, requesterID, defaultExploreLimit)
	if err != nil {
		return nil, err
	}

	threads, err := s.filterPublicThreads(ctx, result.Threads)
	if err != nil {
		return nil, err
	}
	result.Threads = threads
	return result, nil
}

func (s *ExploreService) filterPublicThreads(ctx context.Context, threads []model.ExploreThreadResult) ([]model.ExploreThreadResult, error) {
	userIDs := make([]uint, 0, len(threads))
	seen := make(map[uint]struct{}, len(threads))
	for _, thread := range threads {
		if _, ok := seen[thread.UserID]; ok {
			continue
		}
		seen[thread.UserID] = struct{}{}
		userIDs = append(userIDs, thread.UserID)
	}

	privacyMap, err := s.settingsRepo.GetPrivacyMapByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.ExploreThreadResult, 0, len(threads))
	for _, thread := range threads {
		if !privacyMap[thread.UserID] {
			filtered = append(filtered, thread)
		}
	}
	return filtered, nil
}
