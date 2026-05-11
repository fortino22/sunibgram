package helper

import "github.com/sunibgram/domain/service"

func SendCommentNotification(
	notificationService service.INotificationService,
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
	notificationService service.INotificationService,
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
	_ = notificationService.SendNotification(service.CreateNotificationInput{
		Type:            notificationType,
		Message:         message,
		SenderUserID:    senderUserID,
		RecipientUserID: recipientUserID,
		ReferenceID:     &ref,
	})
}
