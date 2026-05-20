package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

const defaultExploreLimit = 20

type ExploreService struct {
	exploreRepo  repository.IExploreRepository
	threadRepo   repository.IThreadRepository
	settingsRepo repository.IUserSettingsRepository
	followRepo   repository.IFollowRepository
}

func NewExploreService(exploreRepo repository.IExploreRepository, threadRepo repository.IThreadRepository, settingsRepo repository.IUserSettingsRepository, followRepo repository.IFollowRepository) service.IExploreService {
	return &ExploreService{exploreRepo: exploreRepo, threadRepo: threadRepo, settingsRepo: settingsRepo, followRepo: followRepo}
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

func (s *ExploreService) GetTopPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, []*model.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, nil, 0, err
	}
	threads, total, err := s.exploreRepo.FindTopPosts(ctx, query, requesterID, defaultExploreLimit)
	if err != nil {
		return nil, nil, 0, err
	}
	threads, parents, err := s.threadRepo.AttachReplyParents(threads, &requesterID)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
}

func (s *ExploreService) GetLatestPosts(ctx context.Context, query string, requesterID uint) ([]model.Thread, []*model.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, nil, 0, err
	}
	threads, total, err := s.exploreRepo.FindLatestPosts(ctx, query, requesterID, defaultExploreLimit)
	if err != nil {
		return nil, nil, 0, err
	}
	threads, parents, err := s.threadRepo.AttachReplyParents(threads, &requesterID)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
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
	threads, total, err := s.exploreRepo.FindMediaPosts(ctx, query, requesterID, defaultExploreLimit)
	if err != nil {
		return nil, 0, err
	}
	return threads, total, nil
}

func (s *ExploreService) GetCommunities(ctx context.Context, query string, requesterID uint) ([]model.ExploreCommunityResult, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	return s.exploreRepo.FindCommunities(ctx, query, requesterID, defaultExploreLimit)
}

func (s *ExploreService) GetSuggestedPeople(ctx context.Context, requesterID uint, limit int) ([]model.ExploreUserResult, error) {
	_, limit = helper.NormalizePageLimit(1, limit, defaultExploreLimit)
	return s.exploreRepo.FindSuggestedPeople(ctx, requesterID, limit)
}

func normalizeExploreQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", errors.New("query is required")
	}
	return query, nil
}
