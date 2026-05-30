package notificationrepo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	notificationmodel "github.com/sunibgram/domain/model/notification"
	notificationrepoiface "github.com/sunibgram/domain/repository/notification"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

type notificationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewNotificationRepository(writeDB, readDB *gorm.DB, c *cache.Cache) notificationrepoiface.INotificationRepository {
	return &notificationRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func (r *notificationRepository) Save(notification *notificationmodel.Notification) error {
	return r.writeDB.Create(notification).Error
}

func (r *notificationRepository) FindByID(id uint) (*notificationmodel.Notification, error) {
	var notification notificationmodel.Notification
	if err := r.readDB.First(&notification, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) Delete(id uint) error {
	result := r.writeDB.Delete(&notificationmodel.Notification{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}
	return result.Error
}

func (r *notificationRepository) GetNotificationsByRecipientUserID(ctx context.Context, recipientUserID string, limit, offset int) ([]notificationmodel.Notification, error) {
	var notifications []notificationmodel.Notification
	// Drop notifications whose actor is in a mutual block relationship with the
	// recipient (mutual hide). Inline subquery — the notification repo must not
	// depend on the thread repo. recipientUserID is bound as a param; Postgres
	// coerces the text id against the bigint columns.
	err := r.readDB.WithContext(ctx).
		Select("id", "type", "message", "sender_user_id", "reference_id", "created_at").
		Where("recipient_user_id = ?", recipientUserID).
		Where(`sender_user_id NOT IN (
			SELECT blocked_id FROM user_blocks WHERE blocker_id = ?
			UNION
			SELECT blocker_id FROM user_blocks WHERE blocked_id = ?)`, recipientUserID, recipientUserID).
		Where(`sender_user_id NOT IN (
			SELECT muted_user_id FROM notification_user_mutes WHERE user_id = ?)`, recipientUserID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetSenderCache(userID uint) (*notificationmodel.NotificationSender, error) {
	if r.cache == nil {
		return nil, errors.New("cache unavailable")
	}
	var sender notificationmodel.NotificationSender
	if err := r.cache.Get(context.Background(), notificationSenderKey(userID), &sender); err != nil {
		return nil, err
	}
	return &sender, nil
}

func (r *notificationRepository) SetSenderCache(sender *notificationmodel.NotificationSender) error {
	if r.cache == nil || sender == nil {
		return nil
	}
	ctx := context.Background()
	key := notificationSenderKey(sender.UserID)
	if err := r.cache.Set(ctx, key, sender, cache.DefaultTTL); err != nil {
		return err
	}
	// Register under the global tag so infrarepoShared.NotificationSenderGlobalTag-based
	// invalidation (faculty rename, admin user mutations) drops it in
	// O(set size) instead of SCAN.
	_ = r.cache.RegisterCacheKey(ctx, infrarepoShared.NotificationSenderGlobalTag(), key)
	return nil
}
