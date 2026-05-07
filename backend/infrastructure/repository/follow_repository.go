package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type followRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

func NewFollowRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IFollowRepository {
	return &followRepository{
		writeDB: writeDB,
		readDB:  readDB,
		cache:   c,
		ttl:     cache.DefaultTTL,
	}
}

func userStatKey(userID uint) string {
	return fmt.Sprintf("user:stat:%d", userID)
}

func (r *followRepository) Follow(followerID, followingID uint) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	rel := model.UserRelationship{FollowerID: followerID, FollowingID: followingID}
	if err := r.writeDB.Create(&rel).Error; err != nil {
		return errors.New("already following")
	}

	// atomic DB counter update for both users
	r.writeDB.Exec(`
		INSERT INTO user_stats (user_id, follower_count, following_count)
		VALUES (?, 1, 0)
		ON CONFLICT (user_id) DO UPDATE SET follower_count = user_stats.follower_count + 1
	`, followingID)
	r.writeDB.Exec(`
		INSERT INTO user_stats (user_id, follower_count, following_count)
		VALUES (?, 0, 1)
		ON CONFLICT (user_id) DO UPDATE SET following_count = user_stats.following_count + 1
	`, followerID)

	// patch Redis cache for both
	r.patchStatCache(followerID, 0, 1)
	r.patchStatCache(followingID, 1, 0)
	return nil
}

func (r *followRepository) Unfollow(followerID, followingID uint) error {
	result := r.writeDB.Delete(&model.UserRelationship{}, "follower_id = ? AND following_id = ?", followerID, followingID)
	if result.RowsAffected == 0 {
		return errors.New("not following")
	}

	r.writeDB.Exec(`
		UPDATE user_stats SET follower_count = GREATEST(follower_count - 1, 0) WHERE user_id = ?
	`, followingID)
	r.writeDB.Exec(`
		UPDATE user_stats SET following_count = GREATEST(following_count - 1, 0) WHERE user_id = ?
	`, followerID)

	r.patchStatCache(followerID, 0, -1)
	r.patchStatCache(followingID, -1, 0)
	return nil
}

func (r *followRepository) Exists(followerID, followingID uint) (bool, error) {
	var count int64
	r.readDB.Model(&model.UserRelationship{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count)
	return count > 0, nil
}

func (r *followRepository) GetStat(userID uint) (*model.UserStat, error) {
	ctx := context.Background()
	key := userStatKey(userID)
	var stat model.UserStat
	if err := r.cache.Get(ctx, key, &stat); err != nil {
		if err := r.readDB.FirstOrCreate(&stat, model.UserStat{UserID: userID}).Error; err != nil {
			return nil, err
		}
		r.cache.Set(ctx, key, stat, r.ttl)
	}
	return &stat, nil
}

func (r *followRepository) GetFollowing(userID uint, page, limit int) ([]model.User, int64, error) {
	var total int64
	if err := r.readDB.Model(&model.UserRelationship{}).
		Where("follower_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	offset := (page - 1) * limit
	err := r.readDB.Model(&model.User{}).
		Joins("JOIN user_relationships ON user_relationships.following_id = users.user_id").
		Where("user_relationships.follower_id = ?", userID).
		Preload("Role").
		Order("user_relationships.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&users).Error
	return users, total, err
}

func (r *followRepository) GetFollowers(userID uint, page, limit int) ([]model.User, int64, error) {
	var total int64
	if err := r.readDB.Model(&model.UserRelationship{}).
		Where("following_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	offset := (page - 1) * limit
	err := r.readDB.Model(&model.User{}).
		Joins("JOIN user_relationships ON user_relationships.follower_id = users.user_id").
		Where("user_relationships.following_id = ?", userID).
		Preload("Role").
		Order("user_relationships.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&users).Error
	return users, total, err
}

func (r *followRepository) patchStatCache(userID uint, followerDelta, followingDelta int) {
	ctx := context.Background()
	key := userStatKey(userID)
	var stat model.UserStat
	if err := r.cache.Get(ctx, key, &stat); err != nil {
		if err := r.readDB.FirstOrCreate(&stat, model.UserStat{UserID: userID}).Error; err != nil {
			return
		}
	}
	stat.FollowerCount = max(stat.FollowerCount+followerDelta, 0)
	stat.FollowingCount = max(stat.FollowingCount+followingDelta, 0)
	r.cache.Set(ctx, key, stat, r.ttl)
}
