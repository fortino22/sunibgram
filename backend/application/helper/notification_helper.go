package helper

import (
	"log"
	"regexp"
	"strings"

	userrepoiface "github.com/sunibgram/domain/repository/user"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

var mentionPattern = regexp.MustCompile(`@([a-zA-Z0-9_.]+)`)

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

// SendMentionNotifications extracts @username mentions from content and notifies each mentioned user.
// It skips the actor themselves and deduplicates mentions.
func SendMentionNotifications(
	notificationService notificationsvciface.INotificationService,
	userRepo userrepoiface.IUserRepository,
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
		username := strings.ToLower(m[1])
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
