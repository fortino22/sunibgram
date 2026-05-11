package service

import "github.com/sunibgram/domain/model"

type CreateNotificationInput struct {
	Type            string
	Message         string
	SenderUserID    uint
	RecipientUserID uint
	ReferenceID     *uint
}

type CreateNotificationResult struct {
	Notification *model.Notification
	Sender       *model.NotificationSender
}

type INotificationService interface {
	SendNotification(input CreateNotificationInput) (*CreateNotificationResult, error)
	GetNotificationsByRecipientUserID(recipientUserID string, limit, offset int) ([]model.Notification, error)
	DeleteNotification(id, requesterID uint, isAdmin bool) error
}
