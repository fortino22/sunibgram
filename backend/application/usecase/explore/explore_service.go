package exploreuc

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/application/helper"
	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
	explorerepoiface "github.com/sunibgram/domain/repository/explore"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	exploresvciface "github.com/sunibgram/domain/service/explore"
)

const defaultExploreLimit = 20

type ExploreService struct {
	exploreRepo  explorerepoiface.IExploreRepository
	threadRepo   threadrepoiface.IThreadRepository
	settingsRepo usersettingsrepoiface.IUserSettingsRepository
	followRepo   followrepoiface.IFollowRepository
}

func NewExploreService(exploreRepo explorerepoiface.IExploreRepository, threadRepo threadrepoiface.IThreadRepository, settingsRepo usersettingsrepoiface.IUserSettingsRepository, followRepo followrepoiface.IFollowRepository) exploresvciface.IExploreService {
	return &ExploreService{exploreRepo: exploreRepo, threadRepo: threadRepo, settingsRepo: settingsRepo, followRepo: followRepo}
}

func (s *ExploreService) ExploreSearch(ctx context.Context, query string, requesterID *uint) (*exploremodel.ExploreResult, error) {
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

func (s *ExploreService) filterPublicThreads(ctx context.Context, threads []exploremodel.ExploreThreadResult) ([]exploremodel.ExploreThreadResult, error) {
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

	filtered := make([]exploremodel.ExploreThreadResult, 0, len(threads))
	for _, thread := range threads {
		if !privacyMap[thread.UserID] {
			filtered = append(filtered, thread)
		}
	}
	return filtered, nil
}

func (s *ExploreService) GetTopPosts(ctx context.Context, query string, requesterID uint) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
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

func (s *ExploreService) GetLatestPosts(ctx context.Context, query string, requesterID uint) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
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

func (s *ExploreService) GetPeople(ctx context.Context, query string, requesterID uint) ([]exploremodel.ExploreUserResult, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, err
	}
	return s.exploreRepo.FindPeople(ctx, query, requesterID, defaultExploreLimit)
}

func (s *ExploreService) GetMediaPosts(ctx context.Context, query string, requesterID uint) ([]threadmodel.Thread, int64, error) {
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

func (s *ExploreService) GetCommunities(ctx context.Context, query string, requesterID uint) ([]exploremodel.ExploreCommunityResult, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	return s.exploreRepo.FindCommunities(ctx, query, requesterID, defaultExploreLimit)
}

func (s *ExploreService) GetSuggestedPeople(ctx context.Context, requesterID uint, limit int) ([]exploremodel.ExploreUserResult, error) {
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
