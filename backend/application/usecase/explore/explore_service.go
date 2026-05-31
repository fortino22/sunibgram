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

func (s *ExploreService) ExploreSearch(ctx context.Context, query string, requesterID *uint, page, limit int) (*exploremodel.ExploreResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is required")
	}
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)

	result, err := s.exploreRepo.ExploreSearch(ctx, query, requesterID, page, limit)
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

func (s *ExploreService) GetTopPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)
	threads, total, err := s.exploreRepo.FindTopPosts(ctx, query, requesterID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	threads, parents, err := s.threadRepo.AttachReplyParents(threads, &requesterID)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
}

func (s *ExploreService) GetLatestPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)
	threads, total, err := s.exploreRepo.FindLatestPosts(ctx, query, requesterID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	threads, parents, err := s.threadRepo.AttachReplyParents(threads, &requesterID)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
}

func (s *ExploreService) GetPeople(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)
	return s.exploreRepo.FindPeople(ctx, query, requesterID, page, limit)
}

func (s *ExploreService) GetMediaPosts(ctx context.Context, query string, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)
	threads, total, err := s.exploreRepo.FindMediaPosts(ctx, query, requesterID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return threads, total, nil
}

func (s *ExploreService) GetCommunities(ctx context.Context, query string, requesterID uint, page, limit int) ([]exploremodel.ExploreCommunityResult, int64, error) {
	query, err := normalizeExploreQuery(query)
	if err != nil {
		return nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)
	return s.exploreRepo.FindCommunities(ctx, query, requesterID, page, limit)
}

func (s *ExploreService) GetFacultySuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error) {
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)
	return s.exploreRepo.FindFacultySuggestions(ctx, requesterID, page, limit)
}

func (s *ExploreService) GetMutualSuggestions(ctx context.Context, requesterID uint, page, limit int) ([]exploremodel.ExploreUserResult, error) {
	page, limit = helper.NormalizePageLimit(page, limit, defaultExploreLimit)
	return s.exploreRepo.FindMutualSuggestions(ctx, requesterID, page, limit)
}

const suggestedPeoplePreviewLimit = 3

// GetSuggestedPeoplePreview returns up to 3 people for the compact widget,
// cascading sources: mutual (friends-of-friends) → faculty → random.
func (s *ExploreService) GetSuggestedPeoplePreview(ctx context.Context, requesterID uint) ([]exploremodel.ExploreUserResult, error) {
	sources := []func(context.Context, uint, int, int) ([]exploremodel.ExploreUserResult, error){
		s.exploreRepo.FindMutualSuggestions,
		s.exploreRepo.FindFacultySuggestions,
		s.exploreRepo.FindRandomSuggestions,
	}
	for _, find := range sources {
		users, err := find(ctx, requesterID, 1, suggestedPeoplePreviewLimit)
		if err != nil {
			return nil, err
		}
		if len(users) > 0 {
			if len(users) > suggestedPeoplePreviewLimit {
				users = users[:suggestedPeoplePreviewLimit]
			}
			return users, nil
		}
	}
	return []exploremodel.ExploreUserResult{}, nil
}

// GetSuggestedPeople returns up to `limit` people to follow, accumulating
// across sources (mutual → faculty → random) with de-duplication. Unlike the
// preview, it fills from later sources when earlier ones run short — so a brand
// new user with no mutuals/faculty followers still gets a full random list for
// the onboarding gate.
func (s *ExploreService) GetSuggestedPeople(ctx context.Context, requesterID uint, limit int) ([]exploremodel.ExploreUserResult, error) {
	if limit <= 0 || limit > defaultExploreLimit {
		limit = defaultExploreLimit
	}
	sources := []func(context.Context, uint, int, int) ([]exploremodel.ExploreUserResult, error){
		s.exploreRepo.FindMutualSuggestions,
		s.exploreRepo.FindFacultySuggestions,
		s.exploreRepo.FindRandomSuggestions,
	}
	seen := make(map[uint]struct{})
	result := make([]exploremodel.ExploreUserResult, 0, limit)
	for _, find := range sources {
		if len(result) >= limit {
			break
		}
		users, err := find(ctx, requesterID, 1, limit)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			if _, dup := seen[u.UserID]; dup {
				continue
			}
			seen[u.UserID] = struct{}{}
			result = append(result, u)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// onboardingPeopleLimit caps the follow-gate list at a small, scannable size.
const onboardingPeopleLimit = 10

// GetOnboardingPeople backs the first-time follow gate. No query → ranked
// suggestions (already exclude self, admins, already-followed). With a query →
// people search filtered down to non-self results. Either way capped at 10 and
// admin-free, without altering global explore-search behavior.
func (s *ExploreService) GetOnboardingPeople(ctx context.Context, requesterID uint, q string, limit int) ([]exploremodel.ExploreUserResult, error) {
	if limit <= 0 || limit > onboardingPeopleLimit {
		limit = onboardingPeopleLimit
	}

	q = strings.TrimSpace(q)
	if q == "" {
		return s.GetSuggestedPeople(ctx, requesterID, limit)
	}

	// +1 so dropping self in-memory still leaves room for `limit` results.
	people, err := s.GetPeople(ctx, q, requesterID, 1, limit+1)
	if err != nil {
		return nil, err
	}
	result := make([]exploremodel.ExploreUserResult, 0, limit)
	for _, u := range people {
		if u.UserID == requesterID {
			continue
		}
		result = append(result, u)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func normalizeExploreQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", errors.New("query is required")
	}
	return query, nil
}
