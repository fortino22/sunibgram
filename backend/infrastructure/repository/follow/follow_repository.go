package followrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	exploremodel "github.com/sunibgram/domain/model/explore"
	usermodel "github.com/sunibgram/domain/model/user"
	userrelationshipmodel "github.com/sunibgram/domain/model/user_relationship"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	explorerepo "github.com/sunibgram/infrastructure/repository/explore"
)

type followRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

const followingSetTTL = 5 * time.Minute

func NewFollowRepository(writeDB, readDB *gorm.DB, c *cache.Cache) followrepoiface.IFollowRepository {
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

func userFollowingSetKey(userID uint) string {
	return fmt.Sprintf("user:following:set:%d", userID)
}

func (r *followRepository) Follow(followerID, followingID uint) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	rel := userrelationshipmodel.UserRelationship{FollowerID: followerID, FollowingID: followingID}
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
	r.cache.Delete(ctx, userFollowingSetKey(followerID))
	r.invalidateVisibilityFeedsForViewer(ctx, followerID)
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
	return nil
}

func (r *followRepository) Unfollow(followerID, followingID uint) error {
	result := r.writeDB.Delete(&userrelationshipmodel.UserRelationship{}, "follower_id = ? AND following_id = ?", followerID, followingID)
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
	r.cache.Delete(ctx, userFollowingSetKey(followerID))
	r.invalidateVisibilityFeedsForViewer(ctx, followerID)
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
	return nil
}

func (r *followRepository) invalidateVisibilityFeedsForViewer(ctx context.Context, viewerID uint) {
	// Tag-based invalidation. Same blast radius as the legacy 9-pattern SCAN:
	//   infrarepoShared.UserFeedsTag        → thread:userfeed/usermedia/following:<v>:*
	//   infrarepoShared.ViewerThreadFeedsTag → thread:roots:*:viewer:<v>,
	//                          thread:userfeed/usermedia:*:req:<v>,
	//                          thread:my-communities/all-communities:user:<v>:*,
	//                          community:threads:*:user:<v>:*,
	//                          event:threads:*:viewer:<v>:*
	// For You caches use a version-counter (IncrForYouVersion) instead of a tag.
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.UserFeedsTag(viewerID),
		infrarepoShared.ViewerThreadFeedsTag(viewerID),
	)
	infrarepoShared.IncrForYouVersion(ctx, r.cache, viewerID)
}

func (r *followRepository) Exists(followerID, followingID uint) (bool, error) {
	ctx := context.Background()
	key := followExistsKey(followerID, followingID)
	var exists bool
	if err := r.cache.Get(ctx, key, &exists); err == nil {
		return exists, nil
	}
	r.readDB.Model(&userrelationshipmodel.UserRelationship{}).
		Select("1").
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Limit(1).
		Scan(&exists)
	r.cache.Set(ctx, key, exists, r.ttl)
	return exists, nil
}

func (r *followRepository) GetFollowingSet(ctx context.Context, followerID uint, targetIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(targetIDs))
	if len(targetIDs) == 0 {
		return result, nil
	}

	setKey := userFollowingSetKey(followerID)
	targetStrs := make([]string, len(targetIDs))
	for i, id := range targetIDs {
		targetStrs[i] = fmt.Sprintf("%d", id)
	}

	if r.cache != nil {
		if exists, err := r.cache.Exists(ctx, setKey); err == nil && exists {
			if flags, err := r.cache.SMIsMember(ctx, setKey, targetStrs); err == nil {
				for i, ok := range flags {
					if ok {
						result[targetIDs[i]] = true
					}
				}
				return result, nil
			}
		}
	}

	var followingIDs []uint
	if err := r.readDB.WithContext(ctx).
		Model(&userrelationshipmodel.UserRelationship{}).
		Where("follower_id = ?", followerID).
		Pluck("following_id", &followingIDs).Error; err != nil {
		return nil, err
	}

	if r.cache != nil {
		members := make([]string, len(followingIDs))
		for i, id := range followingIDs {
			members[i] = fmt.Sprintf("%d", id)
		}
		_ = r.cache.SAddMany(ctx, setKey, members, followingSetTTL)
	}

	followingSet := make(map[uint]bool, len(followingIDs))
	for _, id := range followingIDs {
		followingSet[id] = true
	}
	for _, id := range targetIDs {
		if followingSet[id] {
			result[id] = true
		}
	}
	return result, nil
}

func (r *followRepository) GetMentionFollowing(ctx context.Context, userID uint, limit int) ([]exploremodel.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = explorerepo.DefaultRepositoryExploreLimit
	}

	var rows []explorerepo.ExploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("user_relationships").
		Select(`
			users.user_id,
			users.name,
			users.user_name AS username,
			users.is_premium,
			COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
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

func (r *followRepository) SearchMentionUsers(ctx context.Context, userID uint, query string, limit int) ([]exploremodel.ExploreUserResult, error) {
	if limit < 1 || limit > 100 {
		limit = explorerepo.DefaultRepositoryExploreLimit
	}

	pattern := explorerepo.LikePattern(query)
	var rows []explorerepo.ExploreUserRow
	if err := r.readDB.WithContext(ctx).
		Table("users").
		Select(`
			users.user_id,
			users.name,
			users.user_name AS username,
			users.is_premium,
			COALESCE(user_profiles.avatar_object_name, '') AS avatar_object_name,
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

func mentionRowsToResults(rows []explorerepo.ExploreUserRow) []exploremodel.ExploreUserResult {
	users := make([]exploremodel.ExploreUserResult, len(rows))
	for i, row := range rows {
		users[i] = exploremodel.ExploreUserResult{
			UserID:           row.UserID,
			Name:             row.Name,
			Username:         row.Username,
			IsPremium:        row.IsPremium,
			AvatarObjectName: row.AvatarObjectName,
			IsFollowing:      row.IsFollowing,
		}
	}
	return users
}

func (r *followRepository) GetStat(userID uint) (*userstatmodel.UserStat, error) {
	ctx := context.Background()
	key := userStatKey(userID)
	var stat userstatmodel.UserStat
	if err := r.cache.Get(ctx, key, &stat); err != nil {
		if err := r.writeDB.FirstOrCreate(&stat, userstatmodel.UserStat{UserID: userID}).Error; err != nil {
			return nil, err
		}
		r.cache.Set(ctx, key, stat, r.ttl)
	}
	return &stat, nil
}

func (r *followRepository) GetFollowing(userID uint, page, limit int) ([]usermodel.User, int64, error) {
	type cached struct {
		Users []usermodel.User
		Total int64
	}

	ctx := context.Background()
	var result cached
	key := followListKey(userID, "following", page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var total int64
		if err := r.readDB.Model(&userrelationshipmodel.UserRelationship{}).
			Where("follower_id = ?", userID).
			Count(&total).Error; err != nil {
			return nil, 0, err
		}

		var users []usermodel.User
		offset := (page - 1) * limit
		if err := r.readDB.Model(&usermodel.User{}).
			Joins("JOIN user_relationships ON user_relationships.following_id = users.user_id").
			Where("user_relationships.follower_id = ?", userID).
			Preload("Role").
			Order("user_relationships.created_at DESC").
			Offset(offset).Limit(limit).
			Find(&users).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Users: users, Total: total}
		if err := r.cache.Set(ctx, key, result, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.FollowListUserTag(userID, "following"), key)
			r.cache.RegisterCacheKey(ctx, infrarepoShared.FollowListGlobalTag(), key)
		}
	}
	return result.Users, result.Total, nil
}

func (r *followRepository) GetFollowers(userID uint, page, limit int) ([]usermodel.User, int64, error) {
	type cached struct {
		Users []usermodel.User
		Total int64
	}

	ctx := context.Background()
	var result cached
	key := followListKey(userID, "followers", page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var total int64
		if err := r.readDB.Model(&userrelationshipmodel.UserRelationship{}).
			Where("following_id = ?", userID).
			Count(&total).Error; err != nil {
			return nil, 0, err
		}

		var users []usermodel.User
		offset := (page - 1) * limit
		if err := r.readDB.Model(&usermodel.User{}).
			Joins("JOIN user_relationships ON user_relationships.follower_id = users.user_id").
			Where("user_relationships.following_id = ?", userID).
			Preload("Role").
			Order("user_relationships.created_at DESC").
			Offset(offset).Limit(limit).
			Find(&users).Error; err != nil {
			return nil, 0, err
		}
		result = cached{Users: users, Total: total}
		if err := r.cache.Set(ctx, key, result, r.ttl); err == nil {
			r.cache.RegisterCacheKey(ctx, infrarepoShared.FollowListUserTag(userID, "followers"), key)
			r.cache.RegisterCacheKey(ctx, infrarepoShared.FollowListGlobalTag(), key)
		}
	}
	return result.Users, result.Total, nil
}

func (r *followRepository) invalidateFollowLists(followerID, followingID uint) {
	ctx := context.Background()
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.FollowListUserTag(followerID, "following"),
		infrarepoShared.FollowListUserTag(followingID, "followers"),
	)
}

func (r *followRepository) patchStatCache(userID uint, followerDelta, followingDelta int) {
	ctx := context.Background()
	key := userStatKey(userID)
	var stat userstatmodel.UserStat
	if err := r.cache.Get(ctx, key, &stat); err != nil {
		if err := r.writeDB.FirstOrCreate(&stat, userstatmodel.UserStat{UserID: userID}).Error; err != nil {
			return
		}
	}
	stat.FollowerCount = max(stat.FollowerCount+followerDelta, 0)
	stat.FollowingCount = max(stat.FollowingCount+followingDelta, 0)
	r.cache.Set(ctx, key, stat, r.ttl)
}
