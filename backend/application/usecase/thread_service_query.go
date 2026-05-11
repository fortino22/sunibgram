package usecase

import (
	"context"
	"errors"

	"github.com/sunibgram/domain/model"
)

func (s *ThreadService) GetByID(id uint) (*model.Thread, error) {
	return s.threadRepo.FindByID(id)
}

func (s *ThreadService) GetRoots(ctx context.Context, requesterID *uint, page, limit int) ([]model.Thread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
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

func (s *ThreadService) GetChain(ctx context.Context, rootID, requesterID uint, page, limit int) ([]model.Thread, int64, error) {
	root, err := s.threadRepo.FindByID(rootID)
	if err != nil {
		return nil, 0, errors.New("thread not found")
	}
	if err := canViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &requesterID, root.UserID); err != nil {
		return nil, 0, err
	}
	if root.ParentID != nil {
		return nil, 0, errors.New("thread is not a root post")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	threads, _, err := s.threadRepo.FindByRootID(rootID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	filtered, err := s.filterViewableThreads(ctx, &requesterID, threads)
	if err != nil {
		return nil, 0, err
	}
	return filtered, int64(len(filtered)), nil
}

func (s *ThreadService) CanViewUserThreads(ctx context.Context, requesterID, targetUserID uint) error {
	return canViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &requesterID, targetUserID)
}

func (s *ThreadService) GetMyPosts(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	threads, parents, _, err := s.threadRepo.FindPostsByUserID(requesterID, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(ctx, requesterID, threads, parents)
}

func (s *ThreadService) GetMyReplies(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
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
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	threads, parents, _, err := s.threadRepo.FindMediaByUserID(requesterID, userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterCommunityMemberFeed(requesterID, threads, parents)
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

func (s *ThreadService) GetCommunityThreads(communityID, userID uint, page, limit int) ([]model.Thread, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return nil, 0, errors.New("forbidden")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.threadRepo.FindByCommunityID(communityID, page, limit)
}

func (s *ThreadService) GetFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	threads, parents, _, err := s.threadRepo.FindFollowingFeed(userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(context.Background(), userID, threads, parents)
}
