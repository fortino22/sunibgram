package profilefieldchangerepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	profilefieldchangemodel "github.com/sunibgram/domain/model/profile_field_change"
	profilefieldchangerepoiface "github.com/sunibgram/domain/repository/profile_field_change"
	"github.com/sunibgram/infrastructure/cache"
)

type ProfileFieldChangeRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func latestKey(userID uint) string {
	return fmt.Sprintf("pfc:user:%d", userID)
}

func historyKey(userID uint, field profilefieldchangemodel.Field) string {
	return fmt.Sprintf("pfc:history:%s:%d", field, userID)
}

// profileHistoryTTL: the change log is append-only and is explicitly invalidated
// in Record on every rename, so it can live far longer than DefaultTTL — the TTL
// is only a backstop. Mirrors OnboardingTTL.
const profileHistoryTTL = 720 * time.Hour // 30 days

func NewProfileFieldChangeRepository(writeDB *gorm.DB, readDB *gorm.DB, c ...*cache.Cache) profilefieldchangerepoiface.IProfileFieldChangeRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &ProfileFieldChangeRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

func (r *ProfileFieldChangeRepository) Record(userID uint, field profilefieldchangemodel.Field, changedAt time.Time, oldValue, newValue string) error {
	rec := &profilefieldchangemodel.ProfileFieldChange{
		UserID:    userID,
		Field:     field,
		ChangedAt: changedAt,
		OldValue:  oldValue,
		NewValue:  newValue,
	}
	if err := r.writeDB.Create(rec).Error; err != nil {
		return err
	}
	if r.cache != nil {
		r.cache.Delete(context.Background(), latestKey(userID), historyKey(userID, field))
	}
	return nil
}

func (r *ProfileFieldChangeRepository) HistoryByUserField(userID uint, field profilefieldchangemodel.Field) ([]profilefieldchangemodel.ProfileFieldChange, error) {
	ctx := context.Background()
	key := historyKey(userID, field)
	if r.cache != nil {
		var cached []profilefieldchangemodel.ProfileFieldChange
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	var rows []profilefieldchangemodel.ProfileFieldChange
	if err := r.readDB.
		Where("user_id = ? AND field = ?", userID, field).
		Order("changed_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		r.cache.Set(ctx, key, rows, profileHistoryTTL)
	}
	return rows, nil
}

func (r *ProfileFieldChangeRepository) LatestByUser(userID uint) (map[profilefieldchangemodel.Field]time.Time, error) {
	ctx := context.Background()
	if r.cache != nil {
		var cached map[profilefieldchangemodel.Field]time.Time
		if err := r.cache.Get(ctx, latestKey(userID), &cached); err == nil {
			return cached, nil
		}
	}
	var rows []struct {
		Field     profilefieldchangemodel.Field
		ChangedAt time.Time
	}
	if err := r.readDB.Model(&profilefieldchangemodel.ProfileFieldChange{}).
		Select("field, MAX(changed_at) as changed_at").
		Where("user_id = ?", userID).
		Group("field").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[profilefieldchangemodel.Field]time.Time, len(rows))
	for _, row := range rows {
		result[row.Field] = row.ChangedAt
	}
	if r.cache != nil {
		r.cache.Set(ctx, latestKey(userID), result, cache.DefaultTTL)
	}
	return result, nil
}
