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
	followRepo   repository.IFollowRepository
}

func NewExploreService(exploreRepo repository.IExploreRepository, settingsRepo repository.IUserSettingsRepository, followRepo repository.IFollowRepository) service.IExploreService {
	return &ExploreService{exploreRepo: exploreRepo, settingsRepo: settingsRepo, followRepo: followRepo}
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

func (s *ExploreService) GetTopPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	threads, _, err := s.exploreRepo.FindTopPosts(ctx, query, defaultExploreLimit)
	if err != nil {
		return nil, 0, err
	}
	return s.filterExploreThreads(ctx, requesterID, threads)
}

func (s *ExploreService) GetLatestPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	threads, _, err := s.exploreRepo.FindLatestPosts(ctx, query, defaultExploreLimit)
	if err != nil {
		return nil, 0, err
	}
	return s.filterExploreThreads(ctx, requesterID, threads)
}

func (s *ExploreService) GetPeople(ctx context.Context, query string, requesterID uint) ([]model.ExploreUserResult, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, err
	}
	return s.exploreRepo.FindPeople(ctx, query, requesterID, defaultExploreLimit)
}

func (s *ExploreService) GetMediaPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	threads, _, err := s.exploreRepo.FindMediaPosts(ctx, query, defaultExploreLimit)
	if err != nil {
		return nil, 0, err
	}
	return s.filterExploreThreads(ctx, requesterID, threads)
}

func (s *ExploreService) GetCommunities(ctx context.Context, query string, requesterID uint) ([]model.ExploreCommunityResult, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	return s.exploreRepo.FindCommunities(ctx, query, requesterID, defaultExploreLimit)
}

func (s *ExploreService) GetSuggestedPeople(ctx context.Context, requesterID uint, limit int) ([]model.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = defaultExploreLimit
	}
	return s.exploreRepo.FindSuggestedPeople(ctx, requesterID, limit)
}

func normalizeExploreQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", errors.New("query is required")
	}
	return query, nil
}

func (s *ExploreService) filterExploreThreads(ctx context.Context, requesterID uint, threads []model.Thread) ([]model.Thread, int64, error) {
	filtered, err := s.filterViewableThreads(ctx, requesterID, threads)
	if err != nil {
		return nil, 0, err
	}
	return filtered, int64(len(filtered)), nil
}

func (s *ExploreService) filterViewableThreads(ctx context.Context, requesterID uint, threads []model.Thread) ([]model.Thread, error) {
	targetIDs := make([]uint, 0, len(threads))
	seen := make(map[uint]struct{}, len(threads))
	for _, thread := range threads {
		if thread.UserID == requesterID {
			continue
		}
		if _, ok := seen[thread.UserID]; ok {
			continue
		}
		seen[thread.UserID] = struct{}{}
		targetIDs = append(targetIDs, thread.UserID)
	}

	privacyMap, err := s.settingsRepo.GetPrivacyMapByUserIDs(ctx, targetIDs)
	if err != nil {
		return nil, err
	}

	privateIDs := make([]uint, 0, len(privacyMap))
	for _, id := range targetIDs {
		if privacyMap[id] {
			privateIDs = append(privateIDs, id)
		}
	}

	followingSet := map[uint]bool{}
	if len(privateIDs) > 0 {
		followingSet, err = s.followRepo.GetFollowingSet(ctx, requesterID, privateIDs)
		if err != nil {
			return nil, err
		}
	}

	filtered := make([]model.Thread, 0, len(threads))
	for _, thread := range threads {
		if thread.UserID == requesterID || !privacyMap[thread.UserID] || followingSet[thread.UserID] {
			filtered = append(filtered, thread)
		}
	}
	return filtered, nil
}
