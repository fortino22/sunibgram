package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
)

func (s *ThreadService) Create(ctx context.Context, userID uint, content string, parentID *uint) (*model.Thread, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is required")
	}

	threadType := model.ThreadTypePost
	var rootID *uint
	var parent *model.Thread

	var communityID *uint

	if parentID != nil {
		foundParent, err := s.threadRepo.FindByID(*parentID)
		if err != nil {
			return nil, errors.New("parent thread not found")
		}
		parent = foundParent
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
	if parent != nil {
		helper.SendCommentNotification(s.notification, userID, parent.UserID, thread.ThreadID)
	}
	return thread, nil
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
	if err := s.threadRepo.Save(repost); err != nil {
		return err
	}
	helper.SendInteractionNotification(s.notification, "repost", "reposted your thread", userID, original.UserID, threadID)
	return nil
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
