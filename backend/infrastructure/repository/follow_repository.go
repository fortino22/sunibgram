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

const followingSetTTL = 5 * time.Minute

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

func followListKey(userID uint, listType string, page, limit int) string {
	return fmt.Sprintf("follows:user:%d:%s:page:%d:limit:%d", userID, listType, page, limit)
}

func followExistsKey(followerID, followingID uint) string {
	return fmt.Sprintf("follow:exists:%d:%d", followerID, followingID)
}

func userFollowingKey(userID uint) string {
	return fmt.Sprintf("user:following:%d", userID)
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
	r.invalidateFollowLists(followerID, followingID)
	ctx := context.Background()
	r.cache.Delete(ctx, followExistsKey(followerID, followingID))
	r.cache.Delete(ctx, userFollowingKey(followerID))
	r.invalidateVisibilityFeedsForViewer(ctx, followerID)
	invalidateExploreCache(ctx, r.cache)
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
	r.invalidateFollowLists(followerID, followingID)
	ctx := context.Background()
	r.cache.Delete(ctx, followExistsKey(followerID, followingID))
	r.cache.Delete(ctx, userFollowingKey(followerID))
	r.invalidateVisibilityFeedsForViewer(ctx, followerID)
	invalidateExploreCache(ctx, r.cache)
	return nil
}

func (r *followRepository) invalidateVisibilityFeedsForViewer(ctx context.Context, viewerID uint) {
	r.cache.DeleteByPatterns(ctx,
		threadForYouFeedPattern(viewerID),
		fmt.Sprintf("thread:following:%d:*", viewerID),
		fmt.Sprintf("thread:roots:*:viewer:%d", viewerID),
		fmt.Sprintf("thread:userfeed:*:req:%d", viewerID),
		fmt.Sprintf("thread:usermedia:*:req:%d", viewerID),
		fmt.Sprintf("thread:my-communities:user:%d:*", viewerID),
		fmt.Sprintf("thread:all-communities:user:%d:*", viewerID),
		fmt.Sprintf("community:threads:*:user:%d:*", viewerID),
		fmt.Sprintf("event:threads:*:viewer:%d:*", viewerID),
	)
}

func (r *followRepository) Exists(followerID, followingID uint) (bool, error) {
	ctx := context.Background()
	key := followExistsKey(followerID, followingID)
	var exists bool
	if err := r.cache.Get(ctx, key, &exists); err == nil {
		return exists, nil
	}
	var count int64
	r.readDB.Model(&model.UserRelationship{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count)
	exists = count > 0
	r.cache.Set(ctx, key, exists, r.ttl)
	return exists, nil
}

func (r *followRepository) GetFollowingSet(ctx context.Context, followerID uint, targetIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(targetIDs))
	if len(targetIDs) == 0 {
		return result, nil
	}

	if r.cache != nil {
		followingSet := map[uint]bool{}
		if err := r.cache.Get(ctx, userFollowingKey(followerID), &followingSet); err == nil {
			for _, id := range targetIDs {
				if followingSet[id] {
					result[id] = true
				}
			}
			return result, nil
		}
	}

	var followingIDs []uint
	if err := r.readDB.WithContext(ctx).
		Model(&model.UserRelationship{}).
		Where("follower_id = ?", followerID).
		Pluck("following_id", &followingIDs).Error; err != nil {
		return nil, err
	}
	followingSet := make(map[uint]bool, len(followingIDs))
	for _, id := range followingIDs {
		followingSet[id] = true
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, userFollowingKey(followerID), followingSet, followingSetTTL)
	}
	for _, id := range targetIDs {
		if followingSet[id] {
			result[id] = true
		}
	}
	return result, nil
}

func (r *followRepository) GetMentionFollowing(ctx context.Context, userID uint, limit int) ([]model.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	var rows []exploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("user_relationships").
		Select(`
			users.user_id,
			users.name,
			users.user_name AS username,
			users.is_premium,
			COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
			COALESCE(user_profiles.avatar_bucket, '') AS avatar_bucket,
			TRUE AS is_following
		`).
		Joins("JOIN users ON users.user_id = user_relationships.following_id").
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Where("user_relationships.follower_id = ?", userID).
		Where("roles.name <> ?", "admin").
		Order("user_relationships.created_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return mentionRowsToResults(rows), nil
}

func (r *followRepository) SearchMentionUsers(ctx context.Context, userID uint, query string, limit int) ([]model.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = defaultRepositoryExploreLimit
	}

	pattern := likePattern(query)
	var rows []exploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("users").
		Select(`
			users.user_id,
			users.name,
			users.user_name AS username,
			users.is_premium,
			COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
			COALESCE(user_profiles.avatar_bucket, '') AS avatar_bucket,
			(user_relationships.following_id IS NOT NULL) AS is_following
		`).
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Joins("LEFT JOIN user_relationships ON user_relationships.following_id = users.user_id AND user_relationships.follower_id = ?", userID).
		Where("roles.name <> ?", "admin").
		Where("users.user_name ILIKE ? ESCAPE '\\' OR users.name ILIKE ? ESCAPE '\\'", pattern, pattern).
		Order("(user_relationships.following_id IS NOT NULL) DESC, users.is_premium DESC, users.created_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return mentionRowsToResults(rows), nil
}

func mentionRowsToResults(rows []exploreUserRow) []model.ExploreUserResult {
	users := make([]model.ExploreUserResult, len(rows))
	for i, row := range rows {
		users[i] = model.ExploreUserResult{
			UserID:           row.UserID,
			Name:             row.Name,
			Username:         row.Username,
			IsPremium:        row.IsPremium,
			AvatarObjectName: row.AvatarObjectName,
			AvatarBucket:     row.AvatarBucket,
			IsFollowing:      row.IsFollowing,
		}
	}
	return users
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
	type cached struct {
		Users []model.User
		Total int64
	}

	ctx := context.Background()
	var result cached
	key := followListKey(userID, "following", page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var total int64
		if err := r.readDB.Model(&model.UserRelationship{}).
			Where("follower_id = ?", userID).
			Count(&total).Error; err != nil {
			return nil, 0, err
		}

		var users []model.User
		offset := (page - 1) * limit
		if err := r.readDB.Model(&model.User{}).
			Joins("JOIN user_relationships ON user_relationships.following_id = users.user_id").
			Where("user_relationships.follower_id = ?", userID).
			Preload("Role").
			Order("user_relationships.created_at DESC").
			Offset(offset).Limit(limit).
			Find(&users).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Users: users, Total: total}
		r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Users, result.Total, nil
}

func (r *followRepository) GetFollowers(userID uint, page, limit int) ([]model.User, int64, error) {
	type cached struct {
		Users []model.User
		Total int64
	}

	ctx := context.Background()
	var result cached
	key := followListKey(userID, "followers", page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var total int64
		if err := r.readDB.Model(&model.UserRelationship{}).
			Where("following_id = ?", userID).
			Count(&total).Error; err != nil {
			return nil, 0, err
		}

		var users []model.User
		offset := (page - 1) * limit
		if err := r.readDB.Model(&model.User{}).
			Joins("JOIN user_relationships ON user_relationships.follower_id = users.user_id").
			Where("user_relationships.following_id = ?", userID).
			Preload("Role").
			Order("user_relationships.created_at DESC").
			Offset(offset).Limit(limit).
			Find(&users).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Users: users, Total: total}
		r.cache.Set(ctx, key, result, r.ttl)
	}
	return result.Users, result.Total, nil
}

func (r *followRepository) invalidateFollowLists(followerID, followingID uint) {
	ctx := context.Background()
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("follows:user:%d:following:*", followerID))
	r.cache.DeleteByPattern(ctx, fmt.Sprintf("follows:user:%d:followers:*", followingID))
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
