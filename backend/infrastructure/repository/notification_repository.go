package repository

import (
	"context"
	"errors"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type notificationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewNotificationRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.INotificationRepository {
	return &notificationRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func (r *notificationRepository) Save(notification *model.Notification) error {
	return r.writeDB.Create(notification).Error
}

func (r *notificationRepository) FindByID(id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := r.readDB.First(&notification, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) Delete(id uint) error {
	result := r.writeDB.Delete(&model.Notification{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}
	return result.Error
}

func (r *notificationRepository) GetNotificationsByRecipientUserID(ctx context.Context, recipientUserID string, limit, offset int) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.readDB.WithContext(ctx).
		Select("id", "type", "message", "sender_user_id", "reference_id", "created_at").
		Where("recipient_user_id = ?", recipientUserID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetSenderCache(userID uint) (*model.NotificationSender, error) {
	if r.cache == nil {
		return nil, errors.New("cache unavailable")
	}
	var sender model.NotificationSender
	if err := r.cache.Get(context.Background(), notificationSenderKey(userID), &sender); err != nil {
		return nil, err
	}
	return &sender, nil
}

func (r *notificationRepository) SetSenderCache(sender *model.NotificationSender) error {
	if r.cache == nil || sender == nil {
		return nil
	}
	return r.cache.Set(context.Background(), notificationSenderKey(sender.UserID), sender, cache.DefaultTTL)
}
