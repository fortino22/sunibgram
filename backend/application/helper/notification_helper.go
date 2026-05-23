package helper

import (
	"log"

	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

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
