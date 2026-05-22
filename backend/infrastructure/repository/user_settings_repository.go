package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *UserSettingsRepository) GetUserSettingByUserID(ctx context.Context, userID uint) (*model.UserSettings, error) {
	var setting model.UserSettings
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
	var setting model.UserSettings
	err := r.readDB.WithContext(ctx).
		Model(&model.UserSettings{}).
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

	var settings []model.UserSettings
	if err := r.readDB.WithContext(ctx).
		Model(&model.UserSettings{}).
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

func (r *UserSettingsRepository) Create(ctx context.Context, setting *model.UserSettings) error {
	return r.writeDB.WithContext(ctx).Create(setting).Error
}

func (r *UserSettingsRepository) UpdateUserSetting(ctx context.Context, setting *model.UserSettings) error {
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
	_ = r.cache.DeleteByPattern(ctx, "thread:userfeed:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:usermedia:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:following:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:roots:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:community:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:all-communities:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:my-communities:*")
	_ = r.cache.DeleteByPattern(ctx, "community:threads:top:*")
	_ = r.cache.DeleteByPattern(ctx, "community:threads:latest:*")
	_ = r.cache.DeleteByPattern(ctx, "feed:foryou:*")
	_ = r.cache.DeleteByPattern(ctx, "event:threads:*")
	_ = r.cache.Delete(ctx, "feed:trending:global")
}
