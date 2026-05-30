package threaduc

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/sunibgram/application/helper"
	communitymodel "github.com/sunibgram/domain/model/community"
	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
)

func (s *ThreadService) GetByID(id uint) (*threadmodel.Thread, error) {
	return s.threadRepo.FindByID(id)
}

func (s *ThreadService) GetRoots(ctx context.Context, requesterID *uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, total, err := s.threadRepo.FindRoots(requesterID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return threads, total, nil
}

func (s *ThreadService) GetForYouFeed(ctx context.Context, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, total, err := s.threadRepo.FindForYouFeed(ctx, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	// For You source queries are not block-filtered at the repo layer; drop
	// blocked-author threads here (total stays approximate, like private filtering).
	threads, parents, err = s.filterBlockedFeed(&userID, threads, parents)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
}

func (s *ThreadService) RefreshForYouFeed(ctx context.Context, userID uint) {
	s.threadRepo.InvalidateForYouSeed(ctx, userID)
}

func (s *ThreadService) CountNewForYouSinceID(ctx context.Context, userID, sinceID uint) (int64, error) {
	return s.threadRepo.CountNewSinceID(ctx, userID, sinceID)
}

func (s *ThreadService) GetNewForYouSinceID(ctx context.Context, userID, sinceID uint, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.threadRepo.FindNewSinceID(ctx, userID, sinceID, limit)
}

func (s *ThreadService) GetChain(ctx context.Context, anchorID, requesterID uint, page, limit int) ([]threadmodel.Thread, int64, error) {
	anchor, ancestors, children, err := s.threadRepo.FindChain(anchorID)
	if err != nil {
		return nil, 0, errors.New("thread not found")
	}

	// For repost threads, redirect the chain to the original thread so replies are visible.
	if anchor.Type == threadmodel.ThreadTypeRepost && anchor.OriginalThreadID != nil {
		anchor, ancestors, children, err = s.threadRepo.FindChain(*anchor.OriginalThreadID)
		if err != nil {
			return nil, 0, errors.New("thread not found")
		}
	}

	if err := CanViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &requesterID, anchor.UserID); err != nil {
		return nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)

	chain := make([]threadmodel.Thread, 0, len(ancestors)+1+len(children))
	chain = append(chain, ancestors...)
	chain = append(chain, anchor)
	chain = append(chain, children...)

	filtered, err := s.filterViewableThreads(ctx, &requesterID, chain)
	if err != nil {
		return nil, 0, err
	}

	// Manual pagination over filtered results
	total := int64(len(filtered))
	offset := (page - 1) * limit
	if offset >= len(filtered) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
}

func (s *ThreadService) CanViewUserThreads(ctx context.Context, requesterID, targetUserID uint) error {
	if s.blockRepo != nil && requesterID != 0 && requesterID != targetUserID {
		blocked, err := s.blockRepo.IsBlocked(requesterID, targetUserID)
		if err != nil {
			return err
		}
		if blocked {
			return errors.New("blocked")
		}
	}
	return CanViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &requesterID, targetUserID)
}

func (s *ThreadService) GetMyPosts(ctx context.Context, requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, _, err := s.threadRepo.FindPostsByUserID(requesterID, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(ctx, requesterID, threads, parents)
}

func (s *ThreadService) GetMyReplies(ctx context.Context, requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, _, err := s.threadRepo.FindReplyTabByUserID(requesterID, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(ctx, requesterID, threads, parents)
}

func (s *ThreadService) GetMyMediaThread(ctx context.Context, requesterID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	if requesterID != userID {
		isPrivate, err := s.settingsRepo.IsUserPrivate(ctx, userID)
		if err != nil {
			return nil, nil, 0, err
		}
		if isPrivate {
			return nil, nil, 0, errors.New("this account is private")
		}
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, _, err := s.threadRepo.FindMediaByUserID(requesterID, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterCommunityMemberFeed(requesterID, threads, parents)
}

func (s *ThreadService) GetMentionUsers(ctx context.Context, userID uint, query string, limit int) ([]exploremodel.ExploreUserResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []exploremodel.ExploreUserResult{}, nil
	}
	return s.followRepo.SearchMentionUsers(ctx, userID, query)
}

func (s *ThreadService) filterUserFeed(ctx context.Context, requesterID uint, threads []threadmodel.Thread, parents []*threadmodel.Thread) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	return threads, parents, int64(len(threads)), nil
}

func (s *ThreadService) filterCommunityMemberFeed(requesterID uint, threads []threadmodel.Thread, parents []*threadmodel.Thread) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	membership := make(map[uint]bool)
	outThreads := make([]threadmodel.Thread, 0, len(threads))
	outParents := make([]*threadmodel.Thread, 0, len(parents))
	for i, thread := range threads {
		if thread.CommunityID != nil {
			allowed, ok := membership[*thread.CommunityID]
			if !ok {
				_, err := s.memberRepo.FindMember(*thread.CommunityID, requesterID)
				allowed = err == nil
				membership[*thread.CommunityID] = allowed
			}
			if !allowed {
				continue
			}
		}
		outThreads = append(outThreads, thread)
		outParents = append(outParents, parents[i])
	}
	return outThreads, outParents, int64(len(outThreads)), nil
}

func (s *ThreadService) ensureCommunityThreadAccess(communityID, userID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if community.Visibility == communitymodel.CommunityVisibilityPublic {
		return nil
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return errors.New("forbidden")
	}
	return nil
}

func (s *ThreadService) GetCommunityThreadTop(communityID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	if err := s.ensureCommunityThreadAccess(communityID, userID); err != nil {
		return nil, nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindCommunityTop(communityID, userID, page, limit)
}

func (s *ThreadService) GetCommunityThreadLatest(communityID, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	if err := s.ensureCommunityThreadAccess(communityID, userID); err != nil {
		return nil, nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindCommunityLatest(communityID, userID, page, limit)
}

func (s *ThreadService) GetCommunityThreadsBySort(communityID, userID uint, sort string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	if err := s.ensureCommunityThreadAccess(communityID, userID); err != nil {
		return nil, nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindCommunityThreadsBySort(communityID, userID, sort, page, limit)
}

func (s *ThreadService) GetCommunityThreadMedia(communityID, userID uint) ([]threadmediamodel.CommunityThreadMedia, error) {
	if err := s.ensureCommunityThreadAccess(communityID, userID); err != nil {
		return nil, err
	}
	return s.threadRepo.FindCommunityMedia(communityID, userID)
}

func (s *ThreadService) GetFollowingFeed(userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, _, err := s.threadRepo.FindFollowingFeed(userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(context.Background(), userID, threads, parents)
}

func (s *ThreadService) GetMyCommunityThreads(ctx context.Context, userID uint, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindMyCommunityThreads(userID, page, limit)
}

func (s *ThreadService) GetAllCommunityThreads(ctx context.Context, userID uint, categoryID *uuid.UUID, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, total, err := s.threadRepo.FindAllCommunityThreads(userID, categoryID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
}

func (s *ThreadService) GetAllCommunityThreadsByQueryTop(ctx context.Context, userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	return s.getAllCommunityThreadsByQuery(ctx, userID, query, page, limit, s.threadRepo.FindAllCommunityThreadsByQueryTop)
}

func (s *ThreadService) GetAllCommunityThreadsByQueryLatest(ctx context.Context, userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	return s.getAllCommunityThreadsByQuery(ctx, userID, query, page, limit, s.threadRepo.FindAllCommunityThreadsByQueryLatest)
}

func (s *ThreadService) getAllCommunityThreadsByQuery(
	ctx context.Context,
	userID uint,
	query string,
	page, limit int,
	load func(userID uint, query string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error),
) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil, 0, errors.New("query is required")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, total, err := load(userID, query, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
}
