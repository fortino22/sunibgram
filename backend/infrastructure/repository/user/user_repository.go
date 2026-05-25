package userrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/infrastructure/cache"

	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	notificationrepo "github.com/sunibgram/infrastructure/repository/notification"
)

type UserRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

func userKey(userID uint) string {
	return fmt.Sprintf("user:id:%d", userID)
}

func NewUserRepository(writeDB *gorm.DB, readDB *gorm.DB, c ...*cache.Cache) userrepoiface.IUserRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &UserRepository{writeDB: writeDB, readDB: readDB, cache: redisCache, ttl: cache.DefaultTTL}
}

func (r *UserRepository) FindByID(id uint) (*usermodel.User, error) {
	ctx := context.Background()
	var user usermodel.User
	key := userKey(id)
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &user); err == nil {
			return &user, nil
		}
	}
	if err := r.readDB.Preload("Role").First(&user, id).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		r.cache.Set(ctx, key, user, r.ttl)
	}
	return &user, nil
}

func (r *UserRepository) FindByIDs(ids []uint) ([]usermodel.User, error) {
	if len(ids) == 0 {
		return []usermodel.User{}, nil
	}
	var users []usermodel.User
	err := r.readDB.Preload("Role").Where("user_id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *UserRepository) Save(user *usermodel.User) error {
	if err := r.writeDB.Create(user).Error; err != nil {
		return err
	}
	notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, user.UserID)
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	return nil
}

func (r *UserRepository) SaveWithProfile(user *usermodel.User, profile *userprofilemodel.UserProfile) error {
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UserID = user.UserID
		return tx.Create(profile).Error
	}); err != nil {
		return err
	}
	notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, user.UserID)
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*usermodel.User, error) {
	var user usermodel.User
	err := r.readDB.Preload("Role").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByUserName(username string) (*usermodel.User, error) {
	var user usermodel.User
	err := r.readDB.Preload("Role").Where("user_name = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user:exists:email:%s", email)
	if r.cache != nil {
		var cached bool
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	var exists bool
	err := r.readDB.Model(&usermodel.User{}).Select("1").Where("email = ?", email).Limit(1).Scan(&exists).Error
	if err == nil && r.cache != nil {
		_ = r.cache.Set(ctx, key, exists, 30*time.Second)
	}
	return exists, err
}

func (r *UserRepository) ExistsByUserName(username string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user:exists:username:%s", username)
	if r.cache != nil {
		var cached bool
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	var exists bool
	err := r.readDB.Model(&usermodel.User{}).Select("1").Where("user_name = ?", username).Limit(1).Scan(&exists).Error
	if err == nil && r.cache != nil {
		_ = r.cache.Set(ctx, key, exists, 30*time.Second)
	}
	return exists, err
}

func (r *UserRepository) UpdatePassword(userID uint, hashedPassword string) error {
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, userID)
	return nil
}

func (r *UserRepository) UpdateInfo(userID uint, name, username string) error {
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id = ?", userID).
		Updates(map[string]any{"name": name, "user_name": username}).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, userID)
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	r.invalidateUserThreadCaches(userID)
	return nil
}

func (r *UserRepository) UpdatePremium(userID uint, premium bool) error {
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id = ?", userID).Update("is_premium", premium).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	r.invalidateProfileMeta(userID)
	notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, userID)
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	r.invalidateUserPremiumCaches()
	r.invalidateUserThreadCaches(userID)
	return nil
}

func (r *UserRepository) UpdatePremiumByIDs(userIDs []uint, premium bool) error {
	if len(userIDs) == 0 {
		return nil
	}
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id IN ?", userIDs).Update("is_premium", premium).Error; err != nil {
		return err
	}
	for _, uid := range userIDs {
		r.invalidateUser(uid)
		r.invalidateProfileMeta(uid)
		notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, uid)
		r.invalidateUserThreadCaches(uid)
	}
	infrarepoShared.InvalidateExploreCache(context.Background(), r.cache)
	r.invalidateUserPremiumCaches()
	return nil
}

// ProfileMetaKey is the Redis key for the FeedBuilder-level profile meta cache.
// Kept in sync with webapiutils.ProfileMetaCacheKey to avoid a cross-package import cycle.
func ProfileMetaKey(userID uint) string {
	return fmt.Sprintf("user:meta:%d", userID)
}

func (r *UserRepository) invalidateProfileMeta(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), ProfileMetaKey(userID))
}

func (r *UserRepository) invalidateUserPremiumCaches() {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, "reports:all")
	// Tag-based invalidation. infrarepoShared.CommunityMembersGlobalTag also covers the
	// community:about:* surface (registered there alongside community:users:*
	// at the Set sites).
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.CommunityMembersGlobalTag(),
		infrarepoShared.CommunityIDGlobalTag(),
	)
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.ReportsIDTag(),
		infrarepoShared.CommunityRequestsGlobalTag(),
	)
	// infrarepoShared.FollowListGlobalTag is registered alongside infrarepoShared.FollowListUserTag at every
	// GetFollowing/GetFollowers Set site — drops every cached follow-list page
	// in O(set size).
	_ = r.cache.InvalidateTag(ctx, infrarepoShared.FollowListGlobalTag())
}

func (r *UserRepository) invalidateUserThreadCaches(userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, "feed:trending:global")
	// For You caches use a version-counter (IncrForYouVersion) instead of a tag.
	infrarepoShared.IncrForYouVersion(ctx, r.cache, userID)
	// Tag-based invalidation.
	//   infrarepoShared.UserFeedsTag           → thread:userfeed/usermedia/following:<uid>:*
	//   infrarepoShared.ViewerThreadFeedsTag   → thread:my-communities:user:<uid>:* + all
	//                            other viewer-scoped feed slices
	//   infrarepoShared.InteractionUserListTag → bookmarks/likes:user:<uid>:*
	//   infrarepoShared.ThreadFollowingGlobalTag / infrarepoShared.ThreadRootsGlobalTag /
	//   infrarepoShared.ThreadCommunityGlobalTag / infrarepoShared.ThreadAllCommunitiesGlobalTag /
	//   infrarepoShared.CommunityThreadsGlobalTag / infrarepoShared.EventThreadsGlobalTag → global feed
	//   families.
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.UserFeedsTag(userID),
		infrarepoShared.ViewerThreadFeedsTag(userID),
		infrarepoShared.InteractionUserListTag("bookmarks", userID),
		infrarepoShared.InteractionUserListTag("likes", userID),
		infrarepoShared.ThreadFollowingGlobalTag(),
		infrarepoShared.ThreadRootsGlobalTag(),
		infrarepoShared.ThreadCommunityGlobalTag(),
		infrarepoShared.ThreadAllCommunitiesGlobalTag(),
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.EventThreadsGlobalTag(),
	)
}

func (r *UserRepository) FindAllPaginated(page, limit int) ([]usermodel.User, int64, error) {
	var users []usermodel.User
	var total int64

	offset := (page - 1) * limit

	if err := r.readDB.Model(&usermodel.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.readDB.Preload("Role").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userKey(userID))
}
