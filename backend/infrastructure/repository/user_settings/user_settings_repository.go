package usersettingsrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
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

func NewUserSettingsRepository(writeDB *gorm.DB, readDB *gorm.DB, c ...*cache.Cache) usersettingsrepoiface.IUserSettingsRepository {
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
	if r.cache == nil {
		missingIDs = userIDs
	} else {
		keys := make([]string, len(userIDs))
		for i, userID := range userIDs {
			keys[i] = userPrivacyKey(userID)
		}
		hits, _ := r.cache.MGet(ctx, keys)
		for _, userID := range userIDs {
			if raw, ok := hits[userPrivacyKey(userID)]; ok {
				var cached bool
				if json.Unmarshal(raw, &cached) == nil {
					result[userID] = cached
					continue
				}
			}
			missingIDs = append(missingIDs, userID)
		}
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
		entries := make(map[string]any, len(missingIDs))
		for _, userID := range missingIDs {
			entries[userPrivacyKey(userID)] = result[userID]
		}
		_ = r.cache.MSet(ctx, entries, r.ttl)
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

func (r *UserSettingsRepository) UpdateUserTheme(ctx context.Context, setting *usersettingsmodel.UserSettings) error {
	// Theme is cosmetic — no privacy/feed cache invalidation needed.
	return r.writeDB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_theme"}),
	}).Create(setting).Error
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
		infrarepoShared.ThreadRootsGlobalTag(),
		infrarepoShared.ThreadCommunityGlobalTag(),
		infrarepoShared.ThreadAllCommunitiesGlobalTag(),
		infrarepoShared.ThreadMyCommunitiesGlobalTag(),
		infrarepoShared.ThreadFollowingGlobalTag(),
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.EventThreadsGlobalTag(),
	)
	// thread:userfeed/usermedia:* still SCAN — no global tag yet (per-user tag
	// is infrarepoShared.UserFeedsTag, registered per-user at each Set; adding
	// a global tag would add a SADD on every user-feed write). Privacy toggle
	// is per-user-rare, SCAN acceptable.
	//
	// feed:foryou:* is also SCAN'd here as a blunt drop. Per-user invalidation
	// via IncrForYouVersion only orphans the current viewer's keys; a privacy
	// toggle affects what every OTHER viewer sees of this author's threads, so
	// the global SCAN is the only way to drop all those viewers' caches.
	_ = r.cache.DeleteByPatterns(ctx,
		"thread:userfeed:*",
		"thread:usermedia:*",
		"feed:foryou:*",
	)
	_ = r.cache.Delete(ctx, "feed:trending:global")
}
