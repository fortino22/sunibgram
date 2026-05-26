package threaduc

import (
	"context"
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/sunibgram/application/helper"
	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

func (s *ThreadService) Create(ctx context.Context, userID uint, content string, parentID *uint, eventID *string, replyPermission *string) (*threadmodel.Thread, error) {
	return s.createThread(ctx, userID, content, parentID, eventID, replyPermission, false)
}

func (s *ThreadService) CreateEventThread(ctx context.Context, userID uint, content string, parentID *uint, eventID *string, replyPermission *string, isOfficial bool) (*threadmodel.Thread, error) {
	if eventID == nil || *eventID == "" {
		return nil, errors.New("event_id is required")
	}
	if isOfficial {
		event, err := s.eventRepo.FindByID(*eventID)
		if err != nil {
			return nil, errors.New("event not found")
		}
		if event.UserID != userID {
			return nil, errors.New("forbidden")
		}
	}
	return s.createThread(ctx, userID, content, parentID, eventID, replyPermission, isOfficial)
}

func (s *ThreadService) createThread(ctx context.Context, userID uint, content string, parentID *uint, eventID *string, replyPermission *string, isOfficial bool) (*threadmodel.Thread, error) {
	content = strings.TrimSpace(content)
	if utf8.RuneCountInString(content) > 5000 {
		return nil, errors.New("content must be 5000 characters or less")
	}
	if err := s.requireEventApproved(eventID); err != nil {
		return nil, err
	}
	if err := s.requireEventJoined(userID, eventID); err != nil {
		return nil, err
	}

	threadType := threadmodel.ThreadTypePost
	var rootID *uint
	var parent *threadmodel.Thread

	var communityID *uint

	resolvedReplyPerm := threadmodel.ReplyPermissionEveryone
	if replyPermission != nil && *replyPermission != "" {
		if !threadmodel.IsValidReplyPermission(*replyPermission) {
			return nil, errors.New("invalid reply_permission")
		}
		if parentID != nil {
			return nil, errors.New("reply_permission can only be set on root threads")
		}
		resolvedReplyPerm = threadmodel.ReplyPermission(*replyPermission)
	}

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
		threadType = threadmodel.ThreadTypeReply
		if parent.RootID == nil {
			rootID = &parent.ThreadID
		} else {
			rootID = parent.RootID
		}
		if err := s.enforceReplyPermission(parent, rootID, userID); err != nil {
			return nil, err
		}
		if parent.CommunityID != nil {
			community, err := s.communityRepo.FindByID(*parent.CommunityID)
			if err != nil {
				return nil, errors.New("community not found")
			}
			if !community.IsWritable() {
				return nil, errors.New("community is inactive")
			}
			if community.Visibility == communitymodel.CommunityVisibilityPrivate {
				if _, err := s.memberRepo.FindMember(*parent.CommunityID, userID); err != nil {
					return nil, errors.New("you must be a member of the community to reply")
				}
			}
			communityID = parent.CommunityID
		}
	}

	thread := &threadmodel.Thread{
		UserID:          userID,
		Type:            threadType,
		Content:         content,
		ReplyPermission: resolvedReplyPerm,
		ParentID:        parentID,
		RootID:          rootID,
		CommunityID:     communityID,
		EventID:         eventID,
		IsOfficial:      isOfficial,
	}
	if err := s.threadRepo.Save(thread); err != nil {
		return nil, err
	}
	if parentID != nil && s.summaryService != nil {
		rootThreadID := *parentID
		if rootID != nil {
			rootThreadID = *rootID
		}
		if err := s.summaryService.RegisterReply(ctx, rootThreadID); err != nil {
			log.Printf("thread summary reply registration failed: %v", err)
		}
	}
	if err := s.persistThreadHashtags(ctx, thread.ThreadID, content, false); err != nil {
		return nil, err
	}
	if parentID != nil {
		counterThreadID := *parentID
		if rootID != nil {
			counterThreadID = *rootID
		}
		if s.hashtagRepo != nil {
			if err := s.hashtagRepo.PatchHashtagCounterByThreadID(ctx, counterThreadID, "thread_count", 1); err != nil {
				log.Printf("reply hashtag trending boost failed: %v", err)
			}
		}
		if s.eventRepo != nil {
			if err := s.eventRepo.PatchEventCounterByThreadID(ctx, counterThreadID, "thread_count", 1); err != nil {
				log.Printf("reply event trending boost failed: %v", err)
			}
		}
	}
	if s.hashtagService != nil {
		s.hashtagService.InvalidateTrendingCache(ctx)
	}
	if parent != nil {
		helper.SendCommentNotification(s.notification, userID, parent.UserID, thread.ThreadID)
	}
	helper.SendMentionNotifications(s.notification, s.userRepo, userID, thread.ThreadID, content)
	return thread, nil
}

func (s *ThreadService) CreateCommunityThread(ctx context.Context, communityID, userID uint, content string, eventID *string) (*threadmodel.Thread, error) {
	content = strings.TrimSpace(content)
	if utf8.RuneCountInString(content) > 5000 {
		return nil, errors.New("content must be 5000 characters or less")
	}
	if err := s.requireEventApproved(eventID); err != nil {
		return nil, err
	}
	if err := s.requireEventJoined(userID, eventID); err != nil {
		return nil, err
	}
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return nil, errors.New("community not found")
	}
	if !community.IsWritable() {
		return nil, errors.New("community is inactive")
	}
	if _, err := s.memberRepo.FindMember(communityID, userID); err != nil {
		return nil, errors.New("forbidden")
	}
	thread := &threadmodel.Thread{
		UserID:      userID,
		CommunityID: &communityID,
		Type:        threadmodel.ThreadTypePost,
		Content:     content,
		EventID:     eventID,
	}
	if err := s.threadRepo.Save(thread); err != nil {
		return nil, err
	}
	if err := s.persistThreadHashtags(ctx, thread.ThreadID, content, false); err != nil {
		return nil, err
	}
	if s.hashtagService != nil {
		s.hashtagService.InvalidateTrendingCache(ctx)
	}
	helper.SendMentionNotifications(s.notification, s.userRepo, userID, thread.ThreadID, content)
	return thread, nil
}

func (s *ThreadService) Update(ctx context.Context, id, userID uint, content string) (*threadmodel.Thread, error) {
	content = strings.TrimSpace(content)
	if utf8.RuneCountInString(content) > 5000 {
		return nil, errors.New("content must be 5000 characters or less")
	}
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("thread not found")
	}
	if thread.UserID != userID {
		return nil, errors.New("forbidden")
	}
	if thread.Type == threadmodel.ThreadTypeRepost {
		return nil, errors.New("cannot edit a repost")
	}
	if err := s.ensureThreadCommunityWritable(thread); err != nil {
		return nil, err
	}
	thread.Content = content
	if err := s.threadRepo.Update(thread); err != nil {
		return nil, err
	}
	if err := s.persistThreadHashtags(ctx, thread.ThreadID, content, true); err != nil {
		return nil, err
	}
	if s.hashtagService != nil {
		s.hashtagService.InvalidateTrendingCache(ctx)
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
	if err := s.ensureThreadCommunityWritable(thread); err != nil {
		return err
	}
	if s.hashtagRepo != nil {
		if err := s.hashtagRepo.UnlinkHashtagsFromThread(ctx, id); err != nil {
			return err
		}
	}
	if err := s.threadRepo.Delete(id); err != nil {
		return err
	}
	if thread.ParentID != nil || thread.RootID != nil {
		counterThreadID := id
		if thread.RootID != nil {
			counterThreadID = *thread.RootID
		} else if thread.ParentID != nil {
			counterThreadID = *thread.ParentID
		}
		if s.hashtagRepo != nil {
			if err := s.hashtagRepo.PatchHashtagCounterByThreadID(ctx, counterThreadID, "thread_count", -1); err != nil {
				log.Printf("reply hashtag trending decrement failed: %v", err)
			}
		}
		if s.eventRepo != nil {
			if err := s.eventRepo.PatchEventCounterByThreadID(ctx, counterThreadID, "thread_count", -1); err != nil {
				log.Printf("reply event trending decrement failed: %v", err)
			}
		}
	}
	if s.hashtagService != nil {
		s.hashtagService.InvalidateTrendingCache(ctx)
	}
	return nil
}

func (s *ThreadService) Repost(ctx context.Context, userID, threadID uint) error {
	original, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if original.Type == threadmodel.ThreadTypeRepost {
		return errors.New("cannot repost a repost")
	}
	if err := CanViewPrivateUser(ctx, s.settingsRepo, s.followRepo, &userID, original.UserID); err != nil {
		return err
	}
	if original.CommunityID != nil {
		community, err := s.communityRepo.FindByID(*original.CommunityID)
		if err != nil {
			return errors.New("community not found")
		}
		if community.Visibility == communitymodel.CommunityVisibilityPrivate {
			if _, err := s.memberRepo.FindMember(*original.CommunityID, userID); err != nil {
				return errors.New("you must be a member of the community to repost this thread")
			}
		}
	}
	exists, err := s.threadRepo.ExistsRepost(userID, threadID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	repost := &threadmodel.Thread{
		UserID:           userID,
		Type:             threadmodel.ThreadTypeRepost,
		Content:          "",
		OriginalThreadID: &original.ThreadID,
	}
	if err := s.threadRepo.Save(repost); err != nil {
		return err
	}
	if s.hashtagRepo != nil {
		if err := s.hashtagRepo.PatchHashtagCounterByThreadID(ctx, threadID, "repost_count", 1); err != nil {
			log.Printf("repost hashtag trending boost failed: %v", err)
		}
	}
	if s.hashtagService != nil {
		s.hashtagService.InvalidateTrendingCache(ctx)
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
		return nil
	}
	if err := s.threadRepo.DeleteRepost(userID, threadID); err != nil {
		return err
	}
	if s.hashtagRepo != nil {
		if err := s.hashtagRepo.PatchHashtagCounterByThreadID(ctx, threadID, "repost_count", -1); err != nil {
			log.Printf("repost hashtag trending decrement failed: %v", err)
		}
	}
	if s.hashtagService != nil {
		s.hashtagService.InvalidateTrendingCache(ctx)
	}
	return nil
}

func (s *ThreadService) IsReposted(userID, threadID uint) bool {
	ok, _ := s.threadRepo.ExistsRepost(userID, threadID)
	return ok
}

func (s *ThreadService) AreThreadsReposted(userID uint, threadIDs []uint) (map[uint]bool, error) {
	return s.threadRepo.AreThreadsReposted(userID, threadIDs)
}

func (s *ThreadService) enforceReplyPermission(parent *threadmodel.Thread, rootID *uint, replierID uint) error {
	root := parent
	if rootID != nil && *rootID != parent.ThreadID {
		fetched, err := s.threadRepo.FindByID(*rootID)
		if err != nil {
			return errors.New("root thread not found")
		}
		root = fetched
	}
	if root.UserID == replierID {
		return nil
	}
	switch root.ReplyPermission {
	case threadmodel.ReplyPermissionEveryone, "":
		return nil
	case threadmodel.ReplyPermissionFollowing:
		authorFollowsReplier, err := s.followRepo.Exists(root.UserID, replierID)
		if err != nil {
			return err
		}
		if !authorFollowsReplier {
			return errors.New("not allowed to reply to this thread")
		}
		return nil
	case threadmodel.ReplyPermissionMutual:
		authorFollowsReplier, err := s.followRepo.Exists(root.UserID, replierID)
		if err != nil {
			return err
		}
		replierFollowsAuthor, err := s.followRepo.Exists(replierID, root.UserID)
		if err != nil {
			return err
		}
		if !authorFollowsReplier || !replierFollowsAuthor {
			return errors.New("not allowed to reply to this thread")
		}
		return nil
	default:
		return errors.New("not allowed to reply to this thread")
	}
}

func (s *ThreadService) UpdateReplyPermission(ctx context.Context, threadID, userID uint, replyPermission string) error {
	if !threadmodel.IsValidReplyPermission(replyPermission) {
		return errors.New("invalid reply_permission")
	}
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return errors.New("thread not found")
	}
	if thread.UserID != userID {
		return errors.New("forbidden")
	}
	if thread.Type != threadmodel.ThreadTypePost {
		return errors.New("reply_permission can only be set on root threads")
	}
	return s.threadRepo.UpdateReplyPermission(threadID, threadmodel.ReplyPermission(replyPermission))
}

func (s *ThreadService) ensureThreadCommunityWritable(thread *threadmodel.Thread) error {
	if thread == nil || thread.CommunityID == nil {
		return nil
	}
	community, err := s.communityRepo.FindByID(*thread.CommunityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	return nil
}

func validateThreadEventID(eventID *string) error {
	if eventID == nil {
		return nil
	}
	if _, err := uuid.Parse(*eventID); err != nil {
		return errors.New("invalid event id")
	}
	return nil
}

func (s *ThreadService) requireEventApproved(eventID *string) error {
	if err := validateThreadEventID(eventID); err != nil {
		return err
	}
	if eventID == nil {
		return nil
	}
	event, err := s.eventRepo.FindByID(*eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status != "approved" {
		return errors.New("event is not approved")
	}
	return nil
}

func (s *ThreadService) requireEventJoined(userID uint, eventID *string) error {
	if eventID == nil || s.eventPartRepo == nil {
		return nil
	}
	joined, err := s.eventPartRepo.Exists(userID, *eventID)
	if err != nil {
		return err
	}
	if !joined {
		return errors.New("you must join the event before posting")
	}
	return nil
}

func (s *ThreadService) persistThreadHashtags(ctx context.Context, threadID uint, content string, replace bool) error {
	if s.hashtagRepo == nil {
		return nil
	}

	names := helper.ExtractHashtags(content)
	if len(names) == 0 {
		if replace {
			return s.hashtagRepo.ReplaceHashtagsForThread(ctx, threadID, nil)
		}
		return nil
	}

	idsByName, err := s.hashtagRepo.UpsertHashtags(ctx, names)
	if err != nil {
		return err
	}
	hashtagIDs := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		if id, ok := idsByName[name]; ok {
			hashtagIDs = append(hashtagIDs, id)
		}
	}
	if replace {
		return s.hashtagRepo.ReplaceHashtagsForThread(ctx, threadID, hashtagIDs)
	}
	return s.hashtagRepo.LinkHashtagsToThread(ctx, threadID, hashtagIDs)
}
