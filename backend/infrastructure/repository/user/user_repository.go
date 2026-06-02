package userrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sunibgram/infrastructure/cache"

	"gorm.io/gorm"

	exploremodel "github.com/sunibgram/domain/model/explore"
	facultymodel "github.com/sunibgram/domain/model/faculty"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	explorerepo "github.com/sunibgram/infrastructure/repository/explore"
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

func onboardingKey(userID uint) string {
	return fmt.Sprintf("user:onboarding:%d", userID)
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

// RegisterAccount persists the full account (user + profile + settings + optional
// faculty assignment) in one writeDB transaction. Either the whole account lands
// or nothing does — no orphaned user row to collide with on retry.
func (r *UserRepository) RegisterAccount(user *usermodel.User, profile *userprofilemodel.UserProfile, settings *usersettingsmodel.UserSettings, facultyID *uint) error {
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UserID = user.UserID
		if err := tx.Create(profile).Error; err != nil {
			return err
		}
		settings.UserID = user.UserID
		if err := tx.Create(settings).Error; err != nil {
			return err
		}
		if facultyID != nil {
			// Omit associations so GORM persists only the join row, not the
			// zero-value embedded User/Faculty structs.
			if err := tx.Omit("User", "Faculty").Create(&usermodel.UserFaculty{UserID: user.UserID, FacultyID: *facultyID}).Error; err != nil {
				return err
			}
			if err := tx.Model(&facultymodel.Faculty{}).
				Where("faculty_id = ?", *facultyID).
				UpdateColumn("total_users", gorm.Expr("total_users + 1")).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		// Map the unique-violation race to a friendly, frontend-mappable message.
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") {
			if strings.Contains(msg, "user_name") || strings.Contains(msg, "username") {
				return errors.New("username already taken")
			}
			return errors.New("email already registered")
		}
		return err
	}

	ctx := context.Background()
	if r.cache != nil {
		_ = r.cache.Delete(ctx,
			fmt.Sprintf("user:exists:email:%s", user.Email),
			fmt.Sprintf("user:exists:username:%s", user.UserName),
		)
	}
	notificationrepo.InvalidateNotificationSenderCache(ctx, r.cache, user.UserID)
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
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
	// Only cache the terminal true state. A cached false is a TOCTOU trap: a
	// registration commits the row but can't cheaply invalidate every pending
	// negative lookup, so a stale false would pass the app check and hit the DB
	// unique constraint. False is cheap to recompute and flips on register.
	if err == nil && exists && r.cache != nil {
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
	// Only cache the terminal true state — see ExistsByEmail for why a cached
	// false is unsafe.
	if err == nil && exists && r.cache != nil {
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

// IsOnboarded reports whether the user finished first-time onboarding. The
// terminal true state is cached long (OnboardingTTL); false is left uncached
// (cheap to recompute, flips soon). Used for an idempotent completion guard.
func (r *UserRepository) IsOnboarded(userID uint) (bool, error) {
	ctx := context.Background()
	key := onboardingKey(userID)
	if r.cache != nil {
		var done bool
		if err := r.cache.Get(ctx, key, &done); err == nil {
			return done, nil
		}
	}
	var user usermodel.User
	if err := r.readDB.Select("onboarding_completed").First(&user, userID).Error; err != nil {
		return false, err
	}
	if r.cache != nil && user.OnboardingCompleted {
		r.cache.Set(ctx, key, true, cache.OnboardingTTL)
	}
	return user.OnboardingCompleted, nil
}

// CountFollowablePeers counts non-admin users excluding the requester — the
// addressable pool for the onboarding follow gate. Mirrors the admin-exclusion
// convention used by the explore suggestion queries.
func (r *UserRepository) CountFollowablePeers(excludeUserID uint) (int64, error) {
	var total int64
	err := r.readDB.
		Table("users").
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Where("roles.name <> ?", "admin").
		Where("users.user_id <> ?", excludeUserID).
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// CompleteOnboarding marks first-time onboarding done. Invalidates the user
// cache so the next profile read reflects it, and primes the long-lived
// onboarding marker.
func (r *UserRepository) CompleteOnboarding(userID uint) error {
	if err := r.writeDB.Model(&usermodel.User{}).Where("user_id = ?", userID).Update("onboarding_completed", true).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	if r.cache != nil {
		r.cache.Set(context.Background(), onboardingKey(userID), true, cache.OnboardingTTL)
	}
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

func (r *UserRepository) SearchForShare(ctx context.Context, query string, excludeUserID uint) ([]exploremodel.ExploreUserResult, error) {
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
			FALSE AS is_following
		`).
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Where("roles.name <> ?", "admin").
		Where("users.user_id <> ?", excludeUserID).
		Where(`users.user_id NOT IN (
			SELECT blocked_id FROM user_blocks WHERE blocker_id = ?
			UNION
			SELECT blocker_id FROM user_blocks WHERE blocked_id = ?)`, excludeUserID, excludeUserID).
		Where(`users.user_id NOT IN (
			SELECT muted_user_id FROM notification_user_mutes WHERE user_id = ?)`, excludeUserID).
		Where("users.user_name ILIKE ? ESCAPE '\\' OR users.name ILIKE ? ESCAPE '\\'", pattern, pattern).
		Order("users.is_premium DESC, users.created_at DESC").
		Limit(10).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	results := make([]exploremodel.ExploreUserResult, len(rows))
	for i, row := range rows {
		results[i] = exploremodel.ExploreUserResult{
			UserID:           row.UserID,
			Name:             row.Name,
			Username:         row.Username,
			IsPremium:        row.IsPremium,
			AvatarObjectName: row.AvatarObjectName,
		}
	}
	return results, nil
}

func (r *UserRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userKey(userID))
}
