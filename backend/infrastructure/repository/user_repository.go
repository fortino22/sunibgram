package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/infrastructure/cache"

	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userrepoiface "github.com/sunibgram/domain/repository/user"
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
	invalidateNotificationSenderCache(context.Background(), r.cache, user.UserID)
	invalidateExploreCache(context.Background(), r.cache)
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
	invalidateNotificationSenderCache(context.Background(), r.cache, user.UserID)
	invalidateExploreCache(context.Background(), r.cache)
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
	var count int64
	err := r.readDB.Model(&usermodel.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByUserName(username string) (bool, error) {
	var count int64
	err := r.readDB.Model(&usermodel.User{}).Where("user_name = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) UpdatePassword(userID uint, hashedPassword string) error {
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	invalidateNotificationSenderCache(context.Background(), r.cache, userID)
	return nil
}

func (r *UserRepository) UpdateInfo(userID uint, name, username string) error {
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id = ?", userID).
		Updates(map[string]any{"name": name, "user_name": username}).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	invalidateNotificationSenderCache(context.Background(), r.cache, userID)
	invalidateExploreCache(context.Background(), r.cache)
	r.invalidateUserThreadCaches(userID)
	return nil
}

func (r *UserRepository) UpdatePremium(userID uint, premium bool) error {
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id = ?", userID).Update("is_premium", premium).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	r.invalidateProfileMeta(userID)
	invalidateNotificationSenderCache(context.Background(), r.cache, userID)
	invalidateExploreCache(context.Background(), r.cache)
	r.invalidateUserPremiumCaches()
	r.invalidateUserThreadCaches(userID)
	return nil
}

// profileMetaKey is the Redis key for the FeedBuilder-level profile meta cache.
// Kept in sync with webapiutils.ProfileMetaCacheKey to avoid a cross-package import cycle.
func profileMetaKey(userID uint) string {
	return fmt.Sprintf("user:meta:%d", userID)
}

func (r *UserRepository) invalidateProfileMeta(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), profileMetaKey(userID))
}

func (r *UserRepository) invalidateUserPremiumCaches() {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, "reports:all")
	// Tag-based invalidation. communityMembersGlobalTag also covers the
	// community:about:* surface (registered there alongside community:users:*
	// at the Set sites).
	_ = r.cache.InvalidateTags(ctx,
		communityThreadsGlobalTag(),
		communityMembersGlobalTag(),
		communityIDGlobalTag(),
	)
	_ = r.cache.InvalidateTags(ctx,
		reportsIDTag(),
		communityRequestsGlobalTag(),
	)
	// followListGlobalTag is registered alongside followListUserTag at every
	// GetFollowing/GetFollowers Set site — drops every cached follow-list page
	// in O(set size).
	_ = r.cache.InvalidateTag(ctx, followListGlobalTag())
}

func (r *UserRepository) invalidateUserThreadCaches(userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, "feed:trending:global")
	// Tag-based invalidation.
	//   userFeedsTag           → thread:userfeed/usermedia/following:<uid>:*
	//   userForYouTag          → feed:foryou:user:<uid>:*
	//   viewerThreadFeedsTag   → thread:my-communities:user:<uid>:* + all
	//                            other viewer-scoped feed slices
	//   interactionUserListTag → bookmarks/likes:user:<uid>:*
	//   threadFollowingGlobalTag / threadRootsGlobalTag /
	//   threadCommunityGlobalTag / threadAllCommunitiesGlobalTag /
	//   communityThreadsGlobalTag / eventThreadsGlobalTag → global feed
	//   families.
	_ = r.cache.InvalidateTags(ctx,
		userFeedsTag(userID),
		userForYouTag(userID),
		viewerThreadFeedsTag(userID),
		interactionUserListTag("bookmarks", userID),
		interactionUserListTag("likes", userID),
		threadFollowingGlobalTag(),
		threadRootsGlobalTag(),
		threadCommunityGlobalTag(),
		threadAllCommunitiesGlobalTag(),
		communityThreadsGlobalTag(),
		eventThreadsGlobalTag(),
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
