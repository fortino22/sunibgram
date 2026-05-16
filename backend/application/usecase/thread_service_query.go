package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
)

func (s *ThreadService) GetByID(id uint) (*model.Thread, error) {
	return s.threadRepo.FindByID(id)
}

func (s *ThreadService) GetRoots(ctx context.Context, requesterID *uint, page, limit int) ([]model.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, _, err := s.threadRepo.FindRoots(page, limit)
	if err != nil {
		return nil, 0, err
	}
	filtered, err := s.filterViewableThreads(ctx, requesterID, threads)
	if err != nil {
		return nil, 0, err
	}
	return filtered, int64(len(filtered)), nil
}

func (s *ThreadService) GetForYouFeed(ctx context.Context, userID uint, page, limit int) ([]model.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	userFaculty, err := s.userFacultyRepo.FindByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	return s.threadRepo.FindForYouFeed(ctx, userID, userFaculty.FacultyID, page, limit)
}

func (s *ThreadService) GetChain(ctx context.Context, anchorID, requesterID uint, page, limit int) ([]model.Thread, int64, error) {
	anchor, err := s.threadRepo.FindByID(anchorID)
	if err != nil {
		return nil, 0, errors.New("thread not found")
	}
	if err := canViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &requesterID, anchor.UserID); err != nil {
		return nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)

	// Determine the true root to fetch all threads in the conversation
	trueRootID := anchorID
	if anchor.RootID != nil {
		trueRootID = *anchor.RootID
	}

	// Fetch all threads in the conversation
	all, _, err := s.threadRepo.FindByRootID(trueRootID, 1, 10000)
	if err != nil {
		return nil, 0, err
	}

	// Collect anchor + only its direct children (parent_id = anchorID)
	var chain []model.Thread
	for _, t := range all {
		if t.ThreadID == anchorID {
			chain = append([]model.Thread{t}, chain...)
			continue
		}
		if t.ParentID != nil && *t.ParentID == anchorID {
			chain = append(chain, t)
		}
	}

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
	return canViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &requesterID, targetUserID)
}

func (s *ThreadService) GetMyPosts(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, _, err := s.threadRepo.FindPostsByUserID(requesterID, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(ctx, requesterID, threads, parents)
}

func (s *ThreadService) GetMyReplies(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, _, err := s.threadRepo.FindReplyTabByUserID(requesterID, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(ctx, requesterID, threads, parents)
}

func (s *ThreadService) GetMyMediaThread(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
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

func (s *ThreadService) GetMentionUsers(ctx context.Context, userID uint, query string, limit int) ([]model.ExploreUserResult, error) {
	_, limit = helper.NormalizePageLimit(1, limit, 20)

	query = strings.TrimSpace(query)
	if query == "" {
		return s.followRepo.GetMentionFollowing(ctx, userID, limit)
	}
	return s.followRepo.SearchMentionUsers(ctx, userID, query, limit)
}

func (s *ThreadService) filterUserFeed(ctx context.Context, requesterID uint, threads []model.Thread, parents []*model.Thread) ([]model.Thread, []*model.Thread, int64, error) {
	filteredThreads, err := s.filterViewableThreads(ctx, &requesterID, threads)
	if err != nil {
		return nil, nil, 0, err
	}
	allowedThreadIDs := make(map[uint]struct{}, len(filteredThreads))
	for _, thread := range filteredThreads {
		allowedThreadIDs[thread.ThreadID] = struct{}{}
	}

	parentThreads := make([]model.Thread, 0, len(parents))
	for _, parent := range parents {
		if parent != nil {
			parentThreads = append(parentThreads, *parent)
		}
	}
	filteredParents, err := s.filterViewableThreads(ctx, &requesterID, parentThreads)
	if err != nil {
		return nil, nil, 0, err
	}
	allowedParentIDs := make(map[uint]struct{}, len(filteredParents))
	for _, parent := range filteredParents {
		allowedParentIDs[parent.ThreadID] = struct{}{}
	}

	outThreads := make([]model.Thread, 0, len(threads))
	outParents := make([]*model.Thread, 0, len(parents))
	for i, thread := range threads {
		if _, ok := allowedThreadIDs[thread.ThreadID]; !ok {
			continue
		}
		parent := parents[i]
		if parent != nil {
			if _, ok := allowedParentIDs[parent.ThreadID]; !ok {
				continue
			}
		}
		outThreads = append(outThreads, thread)
		outParents = append(outParents, parent)
	}
	return outThreads, outParents, int64(len(outThreads)), nil
}

func (s *ThreadService) filterCommunityMemberFeed(requesterID uint, threads []model.Thread, parents []*model.Thread) ([]model.Thread, []*model.Thread, int64, error) {
	membership := make(map[uint]bool)
	outThreads := make([]model.Thread, 0, len(threads))
	outParents := make([]*model.Thread, 0, len(parents))
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
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return errors.New("community not found")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return errors.New("forbidden")
	}
	return nil
}

func (s *ThreadService) GetCommunityThreadTop(communityID, userID uint, page, limit int) ([]model.Thread, int64, error) {
	if err := s.ensureCommunityThreadAccess(communityID, userID); err != nil {
		return nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindCommunityTop(communityID, page, limit)
}

func (s *ThreadService) GetCommunityThreadLatest(communityID, userID uint, page, limit int) ([]model.Thread, int64, error) {
	if err := s.ensureCommunityThreadAccess(communityID, userID); err != nil {
		return nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.threadRepo.FindCommunityLatest(communityID, page, limit)
}

func (s *ThreadService) GetCommunityThreadMedia(communityID, userID uint) ([]model.CommunityThreadMedia, error) {
	if err := s.ensureCommunityThreadAccess(communityID, userID); err != nil {
		return nil, err
	}
	return s.threadRepo.FindCommunityMedia(communityID)
}

func (s *ThreadService) GetFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, parents, _, err := s.threadRepo.FindFollowingFeed(userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(context.Background(), userID, threads, parents)
}
