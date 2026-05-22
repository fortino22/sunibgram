package repository

import (
	"context"

	notificationmodel "github.com/sunibgram/domain/model/notification"
)

type INotificationRepository interface {
	Save(notification *notificationmodel.Notification) error
	FindByID(id uint) (*notificationmodel.Notification, error)
	Delete(id uint) error
	GetNotificationsByRecipientUserID(ctx context.Context, recipientUserID string, limit, offset int) ([]notificationmodel.Notification, error)
	GetSenderCache(userID uint) (*notificationmodel.NotificationSender, error)
	SetSenderCache(sender *notificationmodel.NotificationSender) error
}
