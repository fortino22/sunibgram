package helper

import (
	"log"
	"regexp"
	"strings"

	notificationusermuterepoiface "github.com/sunibgram/domain/repository/notification_user_mute"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

var mentionPattern = regexp.MustCompile(`@\[([a-zA-Z0-9_.]+)\|\d+\]|@([a-zA-Z0-9_.]+)`)

// tagAllPattern matches a bare "@all" mention (case-insensitive, word-bounded)
// used to tag every invitee of a community event.
var tagAllPattern = regexp.MustCompile(`(?i)(^|[^a-zA-Z0-9_.])@all([^a-zA-Z0-9_.]|$)`)

// HasTagAllMention reports whether content contains a bare "@all" mention.
func HasTagAllMention(content string) bool {
	return tagAllPattern.MatchString(content)
}

func SendCommentNotification(
	notificationService notificationsvciface.INotificationService,
	actorUserID uint,
	parentThreadOwnerUserID uint,
	threadID uint,
) {
	SendInteractionNotification(
		notificationService,
		"comment",
		"commented on your thread",
		actorUserID,
		parentThreadOwnerUserID,
		threadID,
	)
}

func SendInteractionNotification(
	notificationService notificationsvciface.INotificationService,
	notificationType string,
	message string,
	senderUserID uint,
	recipientUserID uint,
	referenceID uint,
) {
	if notificationService == nil || senderUserID == recipientUserID {
		return
	}

	ref := referenceID
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("notification helper panic: %v", r)
			}
		}()
		_ = notificationService.SendNotification(notificationsvciface.CreateNotificationInput{
			Type:            notificationType,
			Message:         message,
			SenderUserID:    senderUserID,
			RecipientUserID: recipientUserID,
			ReferenceID:     &ref,
		})
	}()
}

// SendCommunityEventInviteNotifications notifies each tagged user that they were
// added to a community event. Skips the actor and respects block/mute.
func SendCommunityEventInviteNotifications(
	notificationService notificationsvciface.INotificationService,
	blockRepo userblockrepoiface.IUserBlockRepository,
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository,
	actorUserID uint,
	eventID uint,
	recipientUserIDs []uint,
) {
	if notificationService == nil || len(recipientUserIDs) == 0 {
		return
	}
	ref := eventID
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("community event invite notification panic: %v", r)
			}
		}()
		seen := make(map[uint]struct{}, len(recipientUserIDs))
		for _, recipientID := range recipientUserIDs {
			if recipientID == actorUserID {
				continue
			}
			if _, ok := seen[recipientID]; ok {
				continue
			}
			seen[recipientID] = struct{}{}
			if blockRepo != nil {
				if blocked, bErr := blockRepo.IsBlocked(actorUserID, recipientID); bErr == nil && blocked {
					continue
				}
			}
			if muteRepo != nil {
				if muted, mErr := muteRepo.IsMuted(recipientID, actorUserID); mErr == nil && muted {
					continue
				}
			}
			_ = notificationService.SendNotification(notificationsvciface.CreateNotificationInput{
				Type:            "community_event_invite",
				Message:         "tagged you in a community event",
				SenderUserID:    actorUserID,
				RecipientUserID: recipientID,
				ReferenceID:     &ref,
			})
		}
	}()
}

// SendCommunityEventMentionNotifications extracts @username mentions from a
// community event comment and notifies each mentioned user (type
// "community_event_mention", reference = event id).
func SendCommunityEventMentionNotifications(
	notificationService notificationsvciface.INotificationService,
	userRepo userrepoiface.IUserRepository,
	blockRepo userblockrepoiface.IUserBlockRepository,
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository,
	actorUserID uint,
	eventID uint,
	content string,
) {
	if notificationService == nil || userRepo == nil {
		return
	}
	usernames := extractMentionUsernames(content)
	if len(usernames) == 0 {
		return
	}
	ref := eventID
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("community event mention notification panic: %v", r)
			}
		}()
		for _, username := range usernames {
			user, err := userRepo.FindByUserName(username)
			if err != nil {
				continue
			}
			if user.UserID == actorUserID {
				continue
			}
			if blockRepo != nil {
				if blocked, bErr := blockRepo.IsBlocked(actorUserID, user.UserID); bErr == nil && blocked {
					continue
				}
			}
			if muteRepo != nil {
				if muted, mErr := muteRepo.IsMuted(user.UserID, actorUserID); mErr == nil && muted {
					continue
				}
			}
			_ = notificationService.SendNotification(notificationsvciface.CreateNotificationInput{
				Type:            "community_event_mention",
				Message:         "mentioned you in a community event",
				SenderUserID:    actorUserID,
				RecipientUserID: user.UserID,
				ReferenceID:     &ref,
			})
		}
	}()
}

// SendCommunityEventTagAllNotifications notifies every invitee that they were
// @all-mentioned in a community event comment. Skips the actor and respects
// block/mute. Mirrors the mention notification (type "community_event_mention").
func SendCommunityEventTagAllNotifications(
	notificationService notificationsvciface.INotificationService,
	blockRepo userblockrepoiface.IUserBlockRepository,
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository,
	actorUserID uint,
	eventID uint,
	recipientUserIDs []uint,
) {
	if notificationService == nil || len(recipientUserIDs) == 0 {
		return
	}
	ref := eventID
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("community event tag-all notification panic: %v", r)
			}
		}()
		seen := make(map[uint]struct{}, len(recipientUserIDs))
		for _, recipientID := range recipientUserIDs {
			if recipientID == actorUserID {
				continue
			}
			if _, ok := seen[recipientID]; ok {
				continue
			}
			seen[recipientID] = struct{}{}
			if blockRepo != nil {
				if blocked, bErr := blockRepo.IsBlocked(actorUserID, recipientID); bErr == nil && blocked {
					continue
				}
			}
			if muteRepo != nil {
				if muted, mErr := muteRepo.IsMuted(recipientID, actorUserID); mErr == nil && muted {
					continue
				}
			}
			_ = notificationService.SendNotification(notificationsvciface.CreateNotificationInput{
				Type:            "community_event_mention",
				Message:         "mentioned everyone in a community event",
				SenderUserID:    actorUserID,
				RecipientUserID: recipientID,
				ReferenceID:     &ref,
			})
		}
	}()
}

// extractMentionUsernames returns the de-duplicated lowercase usernames found in
// content (both @[username|id] and bare @username forms).
func extractMentionUsernames(content string) []string {
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	usernames := make([]string, 0, len(matches))
	for _, m := range matches {
		raw := m[1]
		if raw == "" {
			raw = m[2]
		}
		username := strings.ToLower(raw)
		if username == "all" {
			// "@all" is the tag-everyone token, not a real username; handled separately.
			continue
		}
		if _, ok := seen[username]; ok {
			continue
		}
		seen[username] = struct{}{}
		usernames = append(usernames, username)
	}
	return usernames
}

// SendMentionNotifications extracts @username mentions from content and notifies each mentioned user.
// It skips the actor themselves and deduplicates mentions.
func SendMentionNotifications(
	notificationService notificationsvciface.INotificationService,
	userRepo userrepoiface.IUserRepository,
	blockRepo userblockrepoiface.IUserBlockRepository,
	muteRepo notificationusermuterepoiface.INotificationUserMuteRepository,
	actorUserID uint,
	threadID uint,
	content string,
) {
	if notificationService == nil || userRepo == nil {
		return
	}

	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return
	}

	seen := make(map[string]struct{}, len(matches))
	usernames := make([]string, 0, len(matches))
	for _, m := range matches {
		raw := m[1]
		if raw == "" {
			raw = m[2]
		}
		username := strings.ToLower(raw)
		if _, ok := seen[username]; ok {
			continue
		}
		seen[username] = struct{}{}
		usernames = append(usernames, username)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("mention notification panic: %v", r)
			}
		}()
		ref := threadID
		for _, username := range usernames {
			user, err := userRepo.FindByUserName(username)
			if err != nil {
				continue
			}
			if user.UserID == actorUserID {
				continue
			}
			// Skip mentioning a user in a mutual block relationship with the actor.
			if blockRepo != nil {
				if blocked, bErr := blockRepo.IsBlocked(actorUserID, user.UserID); bErr == nil && blocked {
					continue
				}
			}
			// Skip mentioning a user who has muted the actor (one-way: recipient muted actor).
			if muteRepo != nil {
				if muted, mErr := muteRepo.IsMuted(user.UserID, actorUserID); mErr == nil && muted {
					continue
				}
			}
			_ = notificationService.SendNotification(notificationsvciface.CreateNotificationInput{
				Type:            "mention",
				Message:         "mentioned you in a thread",
				SenderUserID:    actorUserID,
				RecipientUserID: user.UserID,
				ReferenceID:     &ref,
			})
		}
	}()
}
