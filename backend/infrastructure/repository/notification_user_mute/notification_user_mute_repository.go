package notificationusermuterepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	notificationusermutemodel "github.com/sunibgram/domain/model/notification_user_mute"
	notificationusermuterepoiface "github.com/sunibgram/domain/repository/notification_user_mute"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

type notificationUserMuteRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

func NewNotificationUserMuteRepository(writeDB, readDB *gorm.DB, c *cache.Cache) notificationusermuterepoiface.INotificationUserMuteRepository {
	return &notificationUserMuteRepository{
		writeDB: writeDB,
		readDB:  readDB,
		cache:   c,
		ttl:     cache.DefaultTTL,
	}
}

// muteExistsKey is directional: mute is one-way, so the key is keyed on
// (muter, muted) order — not normalized like the bidirectional block key.
func muteExistsKey(userID, mutedUserID uint) string {
	return fmt.Sprintf("mute:exists:%d:%d", userID, mutedUserID)
}

func (r *notificationUserMuteRepository) Mute(userID, mutedUserID uint) error {
	if userID == mutedUserID {
		return errors.New("cannot mute yourself")
	}

	mute := notificationusermutemodel.NotificationUserMute{UserID: userID, MutedUserID: mutedUserID}
	if err := r.writeDB.Create(&mute).Error; err != nil {
		return errors.New("already muted")
	}

	r.invalidate(userID, mutedUserID)
	return nil
}

func (r *notificationUserMuteRepository) Unmute(userID, mutedUserID uint) error {
	result := r.writeDB.Delete(&notificationusermutemodel.NotificationUserMute{}, "user_id = ? AND muted_user_id = ?", userID, mutedUserID)
	if result.RowsAffected == 0 {
		return errors.New("not muted")
	}

	r.invalidate(userID, mutedUserID)
	return nil
}

// invalidate clears the cached mute flag and the muter's visibility-feed caches.
// Mute is one-way → only the muter's feeds change, so only the muter is invalidated
// (unlike block, which invalidates both sides).
func (r *notificationUserMuteRepository) invalidate(userID, mutedUserID uint) {
	ctx := context.Background()
	r.cache.Delete(ctx, muteExistsKey(userID, mutedUserID))
	r.invalidateVisibilityFeedsForViewer(ctx, userID)
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
}

// invalidateVisibilityFeedsForViewer drops the per-viewer feed caches that bake
// author visibility into their key, so the muted user's posts disappear from the
// muter's feeds on the next read.
func (r *notificationUserMuteRepository) invalidateVisibilityFeedsForViewer(ctx context.Context, viewerID uint) {
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.UserFeedsTag(viewerID),
		infrarepoShared.ViewerThreadFeedsTag(viewerID),
	)
	infrarepoShared.IncrForYouVersion(ctx, r.cache, viewerID)
}

func (r *notificationUserMuteRepository) IsMuted(userID, mutedUserID uint) (bool, error) {
	ctx := context.Background()
	key := muteExistsKey(userID, mutedUserID)
	var exists bool
	if err := r.cache.Get(ctx, key, &exists); err == nil {
		return exists, nil
	}
	if err := r.readDB.Model(&notificationusermutemodel.NotificationUserMute{}).
		Select("1").
		Where("user_id = ? AND muted_user_id = ?", userID, mutedUserID).
		Limit(1).
		Scan(&exists).Error; err != nil {
		return false, err
	}
	r.cache.Set(ctx, key, exists, r.ttl)
	return exists, nil
}

func (r *notificationUserMuteRepository) GetMutedIDs(userID uint) ([]uint, error) {
	var ids []uint
	if err := r.readDB.Model(&notificationusermutemodel.NotificationUserMute{}).
		Where("user_id = ?", userID).
		Pluck("muted_user_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
