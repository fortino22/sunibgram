package communityeventuc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/sunibgram/application/helper"
	storageuc "github.com/sunibgram/application/usecase/storage"
	communitymodel "github.com/sunibgram/domain/model/community"
	communityeventmodel "github.com/sunibgram/domain/model/community_event"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communityeventrepoiface "github.com/sunibgram/domain/repository/community_event"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	notificationusermuterepoiface "github.com/sunibgram/domain/repository/notification_user_mute"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	communityeventsvciface "github.com/sunibgram/domain/service/community_event"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

const communityEventBannerType = "community_event_banner"

type CommunityEventService struct {
	eventRepo      communityeventrepoiface.ICommunityEventRepository
	memberRepo     communitymemberrepoiface.ICommunityMemberRepository
	communityRepo  communityrepoiface.ICommunityRepository
	userRepo       userrepoiface.IUserRepository
	followRepo     followrepoiface.IFollowRepository
	storageUsecase *storageuc.StorageUsecase
	notification   notificationsvciface.INotificationService
	blockRepo      userblockrepoiface.IUserBlockRepository
	muteRepo       notificationusermuterepoiface.INotificationUserMuteRepository
}

func NewCommunityEventService(
	eventRepo communityeventrepoiface.ICommunityEventRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	communityRepo communityrepoiface.ICommunityRepository,
	userRepo userrepoiface.IUserRepository,
	followRepo followrepoiface.IFollowRepository,
	storageUsecase *storageuc.StorageUsecase,
	notification notificationsvciface.INotificationService,
	blockRepo userblockrepoiface.IUserBlockRepository,
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository,
) communityeventsvciface.ICommunityEventService {
	return &CommunityEventService{
		eventRepo:      eventRepo,
		memberRepo:     memberRepo,
		communityRepo:  communityRepo,
		userRepo:       userRepo,
		followRepo:     followRepo,
		storageUsecase: storageUsecase,
		notification:   notification,
		blockRepo:      blockRepo,
		muteRepo:       muteRepo,
	}
}

func (s *CommunityEventService) isAdmin(communityID, userID uint) bool {
	member, err := s.memberRepo.FindMember(communityID, userID)
	if err != nil || member == nil {
		return false
	}
	return member.Role == communitymembermodel.CommunityMemberRoleAdmin
}

func (s *CommunityEventService) canSee(event *communityeventmodel.CommunityEvent, userID uint) bool {
	if s.isAdmin(event.CommunityID, userID) {
		return true
	}
	ok, err := s.eventRepo.IsInvitee(event.ID, userID)
	return err == nil && ok
}

// resolveInvitees maps tagged emails AND user IDs to a deduped user-ID list. For
// PRIVATE communities every invitee must already be a member; for PUBLIC
// communities any existing user is allowed.
func (s *CommunityEventService) resolveInvitees(community *communitymodel.Community, emails []string, userIDs []uint) ([]uint, error) {
	seen := map[uint]bool{}
	out := make([]uint, 0, len(emails)+len(userIDs))
	isPrivate := community.Visibility == communitymodel.CommunityVisibilityPrivate

	add := func(uid uint) {
		if uid == 0 || seen[uid] {
			return
		}
		seen[uid] = true
		out = append(out, uid)
	}

	for _, raw := range emails {
		email := strings.TrimSpace(strings.ToLower(raw))
		if email == "" {
			continue
		}
		user, err := s.userRepo.FindByEmail(email)
		if err != nil || user == nil {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		if isPrivate {
			if _, mErr := s.memberRepo.FindMember(community.CommunityID, user.UserID); mErr != nil {
				return nil, fmt.Errorf("%s is not a member of this community", email)
			}
		}
		add(user.UserID)
	}

	for _, uid := range userIDs {
		if uid == 0 || seen[uid] {
			continue
		}
		user, err := s.userRepo.FindByID(uid)
		if err != nil || user == nil {
			return nil, fmt.Errorf("invalid invitee")
		}
		if isPrivate {
			if _, mErr := s.memberRepo.FindMember(community.CommunityID, uid); mErr != nil {
				return nil, fmt.Errorf("%s is not a member of this community", user.UserName)
			}
		}
		add(uid)
	}

	return out, nil
}

// SearchInviteeCandidates returns taggable users for the picker. PRIVATE
// communities search members only; PUBLIC communities search all users.
func (s *CommunityEventService) SearchInviteeCandidates(ctx context.Context, communityID, requesterID uint, query string) ([]communityeventsvciface.InviteeCandidate, error) {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil || community == nil {
		return nil, errors.New("community not found")
	}
	if !s.isAdmin(communityID, requesterID) {
		return nil, errors.New("forbidden: only community admins can tag people")
	}
	query = strings.TrimSpace(query)

	if community.Visibility == communitymodel.CommunityVisibilityPrivate {
		members, _, mErr := s.memberRepo.FindCommunityUsers(communityID, nil, query, 1, 10)
		if mErr != nil {
			return nil, mErr
		}
		out := make([]communityeventsvciface.InviteeCandidate, 0, len(members))
		for _, m := range members {
			avatar := ""
			if m.Profile != nil {
				avatar = m.Profile.AvatarObjectName
			}
			out = append(out, communityeventsvciface.InviteeCandidate{
				UserID:           m.UserID,
				Username:         m.User.UserName,
				Name:             m.User.Name,
				AvatarObjectName: avatar,
			})
		}
		return out, nil
	}

	results, err := s.followRepo.SearchMentionUsers(ctx, requesterID, query)
	if err != nil {
		return nil, err
	}
	out := make([]communityeventsvciface.InviteeCandidate, 0, len(results))
	for _, r := range results {
		out = append(out, communityeventsvciface.InviteeCandidate{
			UserID:           r.UserID,
			Username:         r.Username,
			Name:             r.Name,
			AvatarObjectName: r.AvatarObjectName,
		})
	}
	return out, nil
}

func validateTextLimits(description, location string) error {
	if utf8.RuneCountInString(strings.TrimSpace(description)) > 2000 {
		return errors.New("description must be 2000 characters or less")
	}
	if utf8.RuneCountInString(strings.TrimSpace(location)) > 255 {
		return errors.New("location must be 255 characters or less")
	}
	return nil
}

func (s *CommunityEventService) Create(ctx context.Context, communityID, actorID uint, input communityeventsvciface.CreateCommunityEventInput) (*communityeventmodel.CommunityEvent, error) {
	if err := communityeventmodel.ValidateCommunityEventInput(input.Title, input.StartTime, input.EndTime); err != nil {
		return nil, err
	}
	if err := validateTextLimits(input.Description, input.Location); err != nil {
		return nil, err
	}
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil || community == nil {
		return nil, errors.New("community not found")
	}
	if !s.isAdmin(communityID, actorID) {
		return nil, errors.New("forbidden: only community admins can create events")
	}
	inviteeIDs, err := s.resolveInvitees(community, input.InviteeEmails, input.InviteeUserIDs)
	if err != nil {
		return nil, err
	}

	event := &communityeventmodel.CommunityEvent{
		CommunityID:     communityID,
		CreatedByUserID: actorID,
		Title:           strings.TrimSpace(input.Title),
		Description:     strings.TrimSpace(input.Description),
		Location:        strings.TrimSpace(input.Location),
		StartTime:       input.StartTime.UTC(),
		EndTime:         input.EndTime.UTC(),
	}
	if input.Banner != nil {
		objName, uErr := helper.UploadImage(ctx, s.storageUsecase, input.Banner, communityEventBannerType)
		if uErr != nil {
			return nil, fmt.Errorf("banner upload failed: %w", uErr)
		}
		event.BannerObjectName = objName
	}
	if err := s.eventRepo.Save(event); err != nil {
		if event.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, event.BannerObjectName)
		}
		return nil, err
	}
	if err := s.eventRepo.ReplaceInvitees(event.ID, inviteeIDs); err != nil {
		return nil, err
	}
	helper.SendCommunityEventInviteNotifications(s.notification, s.blockRepo, s.muteRepo, actorID, event.ID, inviteeIDs)
	return event, nil
}

func (s *CommunityEventService) GetByID(ctx context.Context, eventID, requesterID uint) (*communityeventsvciface.CommunityEventDetail, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("community event not found")
	}
	isAdmin := s.isAdmin(event.CommunityID, requesterID)
	isInvitee, _ := s.eventRepo.IsInvitee(eventID, requesterID)
	if !isAdmin && !isInvitee {
		return nil, errors.New("forbidden: you cannot view this event")
	}
	invitees, err := s.eventRepo.FindInvitees(eventID)
	if err != nil {
		return nil, err
	}
	reactions, err := s.eventRepo.FindReactionSummary(eventID)
	if err != nil {
		return nil, err
	}
	myReaction, _ := s.eventRepo.FindUserReaction(eventID, requesterID)
	var pinned *communityeventmodel.CommunityEventComment
	if event.PinnedCommentID != nil {
		pinned, _ = s.eventRepo.FindCommentByID(*event.PinnedCommentID)
	}
	return &communityeventsvciface.CommunityEventDetail{
		Event:          event,
		Invitees:       invitees,
		ReactionCounts: reactions,
		MyReaction:     myReaction,
		PinnedComment:  pinned,
		IsAdmin:        isAdmin,
		CanComment:     isAdmin || isInvitee,
	}, nil
}

func (s *CommunityEventService) GetByCommunity(ctx context.Context, communityID, requesterID uint, page, limit int) ([]communityeventmodel.CommunityEvent, int64, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, 0, errors.New("community not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	isAdmin := s.isAdmin(communityID, requesterID)
	return s.eventRepo.FindVisibleByCommunity(communityID, requesterID, isAdmin, page, limit)
}

func (s *CommunityEventService) Update(ctx context.Context, eventID, actorID uint, input communityeventsvciface.CreateCommunityEventInput) (*communityeventmodel.CommunityEvent, error) {
	if err := communityeventmodel.ValidateCommunityEventInput(input.Title, input.StartTime, input.EndTime); err != nil {
		return nil, err
	}
	if err := validateTextLimits(input.Description, input.Location); err != nil {
		return nil, err
	}
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("community event not found")
	}
	if !s.isAdmin(event.CommunityID, actorID) {
		return nil, errors.New("forbidden: only community admins can edit events")
	}
	community, err := s.communityRepo.FindByID(event.CommunityID)
	if err != nil || community == nil {
		return nil, errors.New("community not found")
	}

	event.Title = strings.TrimSpace(input.Title)
	event.Description = strings.TrimSpace(input.Description)
	event.Location = strings.TrimSpace(input.Location)
	event.StartTime = input.StartTime.UTC()
	event.EndTime = input.EndTime.UTC()

	if input.ClearBanner && event.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, event.BannerObjectName)
		event.BannerObjectName = ""
	} else if input.Banner != nil {
		if event.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, event.BannerObjectName)
		}
		objName, uErr := helper.UploadImage(ctx, s.storageUsecase, input.Banner, communityEventBannerType)
		if uErr != nil {
			return nil, fmt.Errorf("banner upload failed: %w", uErr)
		}
		event.BannerObjectName = objName
	}

	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}
	if input.InviteeEmails != nil || input.InviteeUserIDs != nil {
		inviteeIDs, rErr := s.resolveInvitees(community, input.InviteeEmails, input.InviteeUserIDs)
		if rErr != nil {
			return nil, rErr
		}
		if err := s.eventRepo.ReplaceInvitees(event.ID, inviteeIDs); err != nil {
			return nil, err
		}
	}
	return event, nil
}

func (s *CommunityEventService) Delete(ctx context.Context, eventID, actorID uint) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("community event not found")
	}
	if !s.isAdmin(event.CommunityID, actorID) {
		return errors.New("forbidden: only community admins can delete events")
	}
	if event.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, event.BannerObjectName)
	}
	return s.eventRepo.Delete(eventID)
}

func (s *CommunityEventService) AddComment(ctx context.Context, eventID, actorID uint, content string) (*communityeventmodel.CommunityEventComment, error) {
	if err := communityeventmodel.ValidateCommentContent(content); err != nil {
		return nil, err
	}
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("community event not found")
	}
	if !s.canSee(event, actorID) {
		return nil, errors.New("forbidden: only tagged people can comment")
	}
	comment := &communityeventmodel.CommunityEventComment{
		CommunityEventID: eventID,
		UserID:           actorID,
		Content:          strings.TrimSpace(content),
	}
	if err := s.eventRepo.SaveComment(comment); err != nil {
		return nil, err
	}
	helper.SendCommunityEventMentionNotifications(s.notification, s.userRepo, s.blockRepo, s.muteRepo, actorID, eventID, comment.Content)
	if helper.HasTagAllMention(comment.Content) {
		if invitees, iErr := s.eventRepo.FindInvitees(eventID); iErr == nil && len(invitees) > 0 {
			ids := make([]uint, 0, len(invitees))
			for _, inv := range invitees {
				ids = append(ids, inv.UserID)
			}
			helper.SendCommunityEventTagAllNotifications(s.notification, s.blockRepo, s.muteRepo, actorID, eventID, ids)
		}
	}
	saved, err := s.eventRepo.FindCommentByID(comment.ID)
	if err != nil {
		return comment, nil
	}
	return saved, nil
}

func (s *CommunityEventService) UpdateComment(ctx context.Context, eventID, commentID, actorID uint, content string) (*communityeventmodel.CommunityEventComment, error) {
	if err := communityeventmodel.ValidateCommentContent(content); err != nil {
		return nil, err
	}
	comment, err := s.eventRepo.FindCommentByID(commentID)
	if err != nil || comment.CommunityEventID != eventID {
		return nil, errors.New("comment not found")
	}
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("community event not found")
	}
	if !s.isAdmin(event.CommunityID, actorID) && comment.UserID != actorID {
		return nil, errors.New("forbidden: you cannot edit this comment")
	}
	comment.Content = strings.TrimSpace(content)
	if err := s.eventRepo.UpdateComment(comment); err != nil {
		return nil, err
	}
	saved, err := s.eventRepo.FindCommentByID(comment.ID)
	if err != nil {
		return comment, nil
	}
	return saved, nil
}

func (s *CommunityEventService) GetComments(ctx context.Context, eventID, requesterID uint, page, limit int) (*communityeventsvciface.CommentList, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, errors.New("community event not found")
	}
	isAdmin := s.isAdmin(event.CommunityID, requesterID)
	isInvitee, _ := s.eventRepo.IsInvitee(eventID, requesterID)
	if !isAdmin && !isInvitee {
		return nil, errors.New("forbidden: you cannot view this event")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 30)
	comments, total, err := s.eventRepo.FindComments(eventID, page, limit)
	if err != nil {
		return nil, err
	}
	return &communityeventsvciface.CommentList{
		Comments:        comments,
		Total:           total,
		PinnedCommentID: event.PinnedCommentID,
		IsAdmin:         isAdmin,
	}, nil
}

func (s *CommunityEventService) DeleteComment(ctx context.Context, eventID, commentID, actorID uint) error {
	comment, err := s.eventRepo.FindCommentByID(commentID)
	if err != nil {
		return errors.New("comment not found")
	}
	if comment.CommunityEventID != eventID {
		return errors.New("comment not found")
	}
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("community event not found")
	}
	if !s.isAdmin(event.CommunityID, actorID) && comment.UserID != actorID {
		return errors.New("forbidden: you cannot delete this comment")
	}
	return s.eventRepo.DeleteComment(commentID)
}

func (s *CommunityEventService) PinComment(ctx context.Context, eventID, commentID, actorID uint) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("community event not found")
	}
	if !s.isAdmin(event.CommunityID, actorID) {
		return errors.New("forbidden: only community admins can pin comments")
	}
	comment, err := s.eventRepo.FindCommentByID(commentID)
	if err != nil || comment.CommunityEventID != eventID {
		return errors.New("comment not found")
	}
	return s.eventRepo.SetPinnedComment(eventID, &commentID)
}

func (s *CommunityEventService) UnpinComment(ctx context.Context, eventID, actorID uint) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("community event not found")
	}
	if !s.isAdmin(event.CommunityID, actorID) {
		return errors.New("forbidden: only community admins can unpin comments")
	}
	return s.eventRepo.SetPinnedComment(eventID, nil)
}

func (s *CommunityEventService) SetReaction(ctx context.Context, eventID, actorID uint, emoji string) (string, error) {
	if err := communityeventmodel.ValidateEmoji(emoji); err != nil {
		return "", err
	}
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return "", errors.New("community event not found")
	}
	if !s.canSee(event, actorID) {
		return "", errors.New("forbidden: you cannot react to this event")
	}
	emoji = strings.TrimSpace(emoji)
	current, _ := s.eventRepo.FindUserReaction(eventID, actorID)
	if current == emoji {
		if err := s.eventRepo.RemoveReaction(eventID, actorID); err != nil {
			return "", err
		}
		return "", nil
	}
	if err := s.eventRepo.UpsertReaction(eventID, actorID, emoji); err != nil {
		return "", err
	}
	return emoji, nil
}

func (s *CommunityEventService) RemoveReaction(ctx context.Context, eventID, actorID uint) error {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return errors.New("community event not found")
	}
	if !s.canSee(event, actorID) {
		return errors.New("forbidden: you cannot react to this event")
	}
	return s.eventRepo.RemoveReaction(eventID, actorID)
}

func (s *CommunityEventService) GetURL(objectName string) string {
	return s.storageUsecase.GetURL(objectName)
}
