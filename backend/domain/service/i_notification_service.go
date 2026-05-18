package service

import "github.com/sunibgram/domain/model"

type CreateNotificationInput struct {
	Type            string
	Message         string
	SenderUserID    uint
	RecipientUserID uint
	ReferenceID     *uint
	Data            any
}

type NotificationListItem struct {
	Notification *model.Notification
	Actor        *model.NotificationSender
	Data         any `json:"data,omitempty"`
}

type INotificationService interface {
	SendNotification(input CreateNotificationInput) error
	GetNotificationsByRecipientUserID(recipientUserID string, limit, offset int) ([]NotificationListItem, error)
	DeleteNotification(id, requesterID uint, isAdmin bool) error
}
