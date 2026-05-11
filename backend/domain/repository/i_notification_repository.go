package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type INotificationRepository interface {
	Save(notification *model.Notification) error
	FindByID(id uint) (*model.Notification, error)
	Delete(id uint) error
	GetNotificationsByRecipientUserID(ctx context.Context, recipientUserID string, limit, offset int) ([]model.Notification, error)
	GetSenderCache(userID uint) (*model.NotificationSender, error)
	SetSenderCache(sender *model.NotificationSender) error
}
