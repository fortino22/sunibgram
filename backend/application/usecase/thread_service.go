package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ThreadService struct {
	threadRepo    repository.IThreadRepository
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	settingsRepo  repository.IUserSettingsRepository
	followRepo    repository.IFollowRepository
}

func NewThreadService(
	threadRepo repository.IThreadRepository,
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	settingsRepo repository.IUserSettingsRepository,
	followRepo repository.IFollowRepository,
) service.IThreadService {
	return &ThreadService{
		threadRepo:    threadRepo,
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		settingsRepo:  settingsRepo,
		followRepo:    followRepo,
	}
}

func (s *ThreadService) Create(ctx context.Context, userID uint, content string, parentID *uint) (*model.Thread, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is required")
	}

	threadType := model.ThreadTypePost
	var rootID *uint

	var communityID *uint

	if parentID != nil {
		parent, err := s.threadRepo.FindByID(*parentID)
		if err != nil {
			return nil, errors.New("parent thread not found")
		}
		if parent.UserID != userID {
			isPrivate, err := s.settingsRepo.IsUserPrivate(ctx, parent.UserID)
			if err != nil {
				return nil, err
			}
			if isPrivate {
				following, err := s.followRepo.Exists(userID, parent.UserID)
				if err != nil {
					return nil, err
				}
				if !following {
					return nil, errors.New("this account is private")
				}
			}
		}
		threadType = model.ThreadTypeReply
		if parent.RootID == nil {
			rootID = &parent.ThreadID
		} else {
			rootID = parent.RootID
		}
		if parent.CommunityID != nil {
			if _, err := s.memberRepo.FindMember(*parent.CommunityID, userID); err != nil {
				return nil, errors.New("you must be a member of the community to reply")
			}
			communityID = parent.CommunityID
		}
	}

	thread := &model.Thread{
		UserID:      userID,
		Type:        threadType,
		Content:     content,
		ParentID:    parentID,
		RootID:      rootID,
		CommunityID: communityID,
	}
	if err := s.threadRepo.Save(thread); err != nil {
		return nil, err
	}
	return thread, nil
}

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

func (s *ThreadService) Update(ctx context.Context, id, userID uint, content string) (*model.Thread, error) {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("thread not found")
	}
	if thread.UserID != userID {
		return nil, errors.New("forbidden")
	}
	if thread.Type == model.ThreadTypeRepost {
		return nil, errors.New("cannot edit a repost")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is required")
	}
	thread.Content = content
	if err := s.threadRepo.Update(thread); err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *ThreadService) Delete(ctx context.Context, id, userID uint) error {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}
	return s.threadRepo.Delete(id)
}

func (s *ThreadService) Repost(ctx context.Context, userID, threadID uint) error {
	original, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if original.Type == model.ThreadTypeRepost {
		return errors.New("cannot repost a repost")
	}
	if err := canViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &userID, original.UserID); err != nil {
		return err
	}
	if original.CommunityID != nil {
		if _, err := s.memberRepo.FindMember(*original.CommunityID, userID); err != nil {
			return errors.New("you must be a member of the community to repost this thread")
		}
	}
	exists, err := s.threadRepo.ExistsRepost(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already reposted")
	}
	repost := &model.Thread{
		UserID:           userID,
		Type:             model.ThreadTypeRepost,
		Content:          "",
		OriginalThreadID: &original.ThreadID,
	}
	return s.threadRepo.Save(repost)
}

func (s *ThreadService) Unrepost(ctx context.Context, userID, threadID uint) error {
	exists, err := s.threadRepo.ExistsRepost(userID, threadID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("not reposted")
	}
	return s.threadRepo.DeleteRepost(userID, threadID)
}

func (s *ThreadService) IsReposted(userID, threadID uint) bool {
	ok, _ := s.threadRepo.ExistsRepost(userID, threadID)
	return ok
}

func (s *ThreadService) CanViewUserThreads(ctx context.Context, requesterID, targetUserID uint) error {
	return canViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &requesterID, targetUserID)
}

func (s *ThreadService) filterViewableThreads(ctx context.Context, requesterID *uint, threads []model.Thread) ([]model.Thread, error) {
	targetIDs := make([]uint, 0, len(threads))
	seen := make(map[uint]struct{}, len(threads))
	for _, thread := range threads {
		if requesterID != nil && thread.UserID == *requesterID {
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
	if requesterID != nil {
		followingSet, err = s.followRepo.GetFollowingSet(ctx, *requesterID, privateIDs)
		if err != nil {
			return nil, err
		}
	}

	filtered := make([]model.Thread, 0, len(threads))
	for _, thread := range threads {
		if requesterID != nil && thread.UserID == *requesterID {
			filtered = append(filtered, thread)
			continue
		}
		if !privacyMap[thread.UserID] || followingSet[thread.UserID] {
			filtered = append(filtered, thread)
		}
	}
	return filtered, nil
}

func canViewPrivateUser(
	ctx context.Context,
	settingsRepo repository.IUserSettingsRepository,
	followRepo repository.IFollowRepository,
	requesterID *uint,
	targetUserID uint,
) error {
	if requesterID != nil && *requesterID == targetUserID {
		return nil
	}
	isPrivate, err := settingsRepo.IsUserPrivate(ctx, targetUserID)
	if err != nil {
		return err
	}
	if !isPrivate {
		return nil
	}
	if requesterID == nil {
		return errors.New("this account is private")
	}
	following, err := followRepo.Exists(*requesterID, targetUserID)
	if err != nil {
		return err
	}
	if !following {
		return errors.New("this account is private")
	}
	return nil
}

func (s *ThreadService) GetMyPosts(ctx context.Context, requesterID, userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	threads, parents, _, err := s.threadRepo.FindPostsByUserID(userID, page, limit)
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
	threads, parents, _, err := s.threadRepo.FindReplyTabByUserID(userID, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	return s.filterUserFeed(ctx, requesterID, threads, parents)
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

func (s *ThreadService) CreateCommunityThread(ctx context.Context, communityID, userID uint, content string) (*model.Thread, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is required")
	}
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, errors.New("community not found")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return nil, errors.New("forbidden")
	}
	thread := &model.Thread{
		UserID:      userID,
		CommunityID: &communityID,
		Type:        model.ThreadTypePost,
		Content:     content,
	}
	if err := s.threadRepo.Save(thread); err != nil {
		return nil, err
	}
	return thread, nil
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
	return s.threadRepo.FindFollowingFeed(userID, page, limit)
}
