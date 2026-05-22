package service

import notificationmodel "github.com/sunibgram/domain/model/notification"

type CreateNotificationInput struct {
	Type            string
	Message         string
	SenderUserID    uint
	RecipientUserID uint
	ReferenceID     *uint
	Data            any
}

type NotificationListItem struct {
	Notification *notificationmodel.Notification
	Actor        *notificationmodel.NotificationSender
	Data         any `json:"data,omitempty"`
}

type INotificationService interface {
	SendNotification(input CreateNotificationInput) error
	GetNotificationsByRecipientUserID(recipientUserID string, limit, offset int) ([]NotificationListItem, error)
	DeleteNotification(id, requesterID uint, isAdmin bool) error
}
