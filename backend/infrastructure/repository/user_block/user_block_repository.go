package userblockrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	userblockmodel "github.com/sunibgram/domain/model/user_block"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

type userBlockRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

func NewUserBlockRepository(writeDB, readDB *gorm.DB, c *cache.Cache) userblockrepoiface.IUserBlockRepository {
	return &userBlockRepository{
		writeDB: writeDB,
		readDB:  readDB,
		cache:   c,
		ttl:     cache.DefaultTTL,
	}
}

func blockExistsKey(viewerID, targetID uint) string {
	lo, hi := viewerID, targetID
	if lo > hi {
		lo, hi = hi, lo
	}
	return fmt.Sprintf("block:exists:%d:%d", lo, hi)
}

func (r *userBlockRepository) Block(blockerID, blockedID uint) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	block := userblockmodel.UserBlock{BlockerID: blockerID, BlockedID: blockedID}
	if err := r.writeDB.Create(&block).Error; err != nil {
		return errors.New("already blocked")
	}

	r.invalidate(blockerID, blockedID)
	return nil
}

func (r *userBlockRepository) Unblock(blockerID, blockedID uint) error {
	result := r.writeDB.Delete(&userblockmodel.UserBlock{}, "blocker_id = ? AND blocked_id = ?", blockerID, blockedID)
	if result.RowsAffected == 0 {
		return errors.New("not blocked")
	}

	r.invalidate(blockerID, blockedID)
	return nil
}

// invalidate clears the cached block flag and both users' visibility-feed caches.
// Mutual hide → both viewers' feeds change, so both are invalidated.
func (r *userBlockRepository) invalidate(blockerID, blockedID uint) {
	ctx := context.Background()
	r.cache.Delete(ctx, blockExistsKey(blockerID, blockedID))
	r.invalidateVisibilityFeedsForViewer(ctx, blockerID)
	r.invalidateVisibilityFeedsForViewer(ctx, blockedID)
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
}

// invalidateVisibilityFeedsForViewer mirrors the follow repository helper: it drops
// the per-viewer feed caches that bake author visibility into their key.
func (r *userBlockRepository) invalidateVisibilityFeedsForViewer(ctx context.Context, viewerID uint) {
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.UserFeedsTag(viewerID),
		infrarepoShared.ViewerThreadFeedsTag(viewerID),
	)
	infrarepoShared.IncrForYouVersion(ctx, r.cache, viewerID)
}

func (r *userBlockRepository) IsBlocked(viewerID, targetID uint) (bool, error) {
	ctx := context.Background()
	key := blockExistsKey(viewerID, targetID)
	var exists bool
	if err := r.cache.Get(ctx, key, &exists); err == nil {
		return exists, nil
	}
	if err := r.readDB.Model(&userblockmodel.UserBlock{}).
		Select("1").
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
			viewerID, targetID, targetID, viewerID).
		Limit(1).
		Scan(&exists).Error; err != nil {
		return false, err
	}
	r.cache.Set(ctx, key, exists, r.ttl)
	return exists, nil
}

func (r *userBlockRepository) BlocksUser(blockerID, blockedID uint) (bool, error) {
	var exists bool
	if err := r.readDB.Model(&userblockmodel.UserBlock{}).
		Select("1").
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Limit(1).
		Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (r *userBlockRepository) GetBlockedIDs(userID uint) ([]uint, error) {
	var ids []uint
	if err := r.readDB.Model(&userblockmodel.UserBlock{}).
		Where("blocker_id = ?", userID).
		Pluck("blocked_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
