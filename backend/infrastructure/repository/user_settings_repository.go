package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
)

type UserSettingsRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

const privacySettingsTTL = 5 * time.Minute

func userPrivacyKey(userID uint) string {
	return fmt.Sprintf("user:privacy:%d", userID)
}

func NewUserSettingsRepository(writeDB *gorm.DB, readDB *gorm.DB, c ...*cache.Cache) domainrepo.IUserSettingsRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &UserSettingsRepository{writeDB: writeDB, readDB: readDB, cache: redisCache, ttl: privacySettingsTTL}
}

func (r *UserSettingsRepository) GetUserSettingByUserID(ctx context.Context, userID uint) (*usersettingsmodel.UserSettings, error) {
	var setting usersettingsmodel.UserSettings
	err := r.readDB.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error
	return &setting, err
}

func (r *UserSettingsRepository) IsUserPrivate(ctx context.Context, userID uint) (bool, error) {
	if r.cache != nil {
		var cached bool
		if err := r.cache.Get(ctx, userPrivacyKey(userID), &cached); err == nil {
			return cached, nil
		}
	}
	var setting usersettingsmodel.UserSettings
	err := r.readDB.WithContext(ctx).
		Model(&usersettingsmodel.UserSettings{}).
		Select("is_private").
		Where("user_id = ?", userID).
		Scan(&setting).Error
	if err == nil && r.cache != nil {
		r.cache.Set(ctx, userPrivacyKey(userID), setting.IsPrivate, r.ttl)
	}
	return setting.IsPrivate, err
}

func (r *UserSettingsRepository) GetPrivacyMapByUserIDs(ctx context.Context, userIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	missingIDs := make([]uint, 0, len(userIDs))
	for _, userID := range userIDs {
		if r.cache == nil {
			missingIDs = append(missingIDs, userID)
			continue
		}
		var cached bool
		if err := r.cache.Get(ctx, userPrivacyKey(userID), &cached); err == nil {
			result[userID] = cached
			continue
		}
		missingIDs = append(missingIDs, userID)
	}
	if len(missingIDs) == 0 {
		return result, nil
	}

	var settings []usersettingsmodel.UserSettings
	if err := r.readDB.WithContext(ctx).
		Model(&usersettingsmodel.UserSettings{}).
		Select("user_id", "is_private").
		Where("user_id IN ?", missingIDs).
		Find(&settings).Error; err != nil {
		return nil, err
	}
	for _, setting := range settings {
		result[setting.UserID] = setting.IsPrivate
	}
	if r.cache != nil {
		for _, userID := range missingIDs {
			r.cache.Set(ctx, userPrivacyKey(userID), result[userID], r.ttl)
		}
	}
	return result, nil
}

func (r *UserSettingsRepository) Create(ctx context.Context, setting *usersettingsmodel.UserSettings) error {
	return r.writeDB.WithContext(ctx).Create(setting).Error
}

func (r *UserSettingsRepository) UpdateUserSetting(ctx context.Context, setting *usersettingsmodel.UserSettings) error {
	if err := r.writeDB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_private"}),
	}).Create(setting).Error; err != nil {
		return err
	}
	r.invalidatePrivacy(ctx, setting.UserID)
	return nil
}

func (r *UserSettingsRepository) invalidatePrivacy(ctx context.Context, userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(ctx, userPrivacyKey(userID))
	r.invalidateVisibilityDependentCaches(ctx)
}

func (r *UserSettingsRepository) invalidateVisibilityDependentCaches(ctx context.Context) {
	if r.cache == nil {
		return
	}
	// Per-family global tags cover thread:roots/community/all-communities/
	// my-communities/following + community:threads + event:threads at every
	// Set site.
	_ = r.cache.InvalidateTags(ctx,
		threadRootsGlobalTag(),
		threadCommunityGlobalTag(),
		threadAllCommunitiesGlobalTag(),
		threadMyCommunitiesGlobalTag(),
		threadFollowingGlobalTag(),
		communityThreadsGlobalTag(),
		eventThreadsGlobalTag(),
	)
	// thread:userfeed/usermedia:* and feed:foryou:* still SCAN — no global
	// tag yet (per-user tags are userFeedsTag / userForYouTag, registered
	// per-user at each Set; adding global tags would add a SADD on every
	// user-feed write). Privacy toggle is per-user-rare, SCAN acceptable.
	_ = r.cache.DeleteByPatterns(ctx,
		"thread:userfeed:*",
		"thread:usermedia:*",
		"feed:foryou:*",
	)
	_ = r.cache.Delete(ctx, "feed:trending:global")
}
