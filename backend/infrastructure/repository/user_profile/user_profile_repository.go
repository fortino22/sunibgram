package userprofilerepo

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userprofilerepoiface "github.com/sunibgram/domain/repository/user_profile"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	notificationrepo "github.com/sunibgram/infrastructure/repository/notification"
)

type UserProfileRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewUserProfileRepository(writeDB *gorm.DB, readDB *gorm.DB, c ...*cache.Cache) userprofilerepoiface.IUserProfileRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &UserProfileRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

func userProfileKey(userID uint) string {
	return fmt.Sprintf("profile:user:%d", userID)
}

func (r *UserProfileRepository) Upsert(profile *userprofilemodel.UserProfile) error {
	if profile.ProfileID != 0 {
		if err := r.writeDB.Model(profile).Updates(map[string]any{
			"avatar_object_name": profile.AvatarObjectName,
			"banner_object_name": profile.BannerObjectName,
			"bio":                profile.Bio,
			"gender":             profile.Gender,
			"birthday_date":      profile.BirthdayDate,
		}).Error; err != nil {
			return err
		}
		r.invalidateUser(profile.UserID)
		ctx := context.Background()
		notificationrepo.InvalidateNotificationSenderCache(ctx, r.cache, profile.UserID)
		infrarepoShared.InvalidateExploreCache(ctx, r.cache)
		if r.cache != nil {
			_ = r.cache.InvalidateTags(ctx,
				infrarepoShared.CommunityMembersGlobalTag(),
				infrarepoShared.CommunityIDGlobalTag(),
			)
		}
		return nil
	}
	if err := r.writeDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"avatar_object_name", "banner_object_name", "bio", "gender", "birthday_date", "updated_at"}),
	}).Create(profile).Error; err != nil {
		return err
	}
	r.invalidateUser(profile.UserID)
	ctx := context.Background()
	notificationrepo.InvalidateNotificationSenderCache(ctx, r.cache, profile.UserID)
	infrarepoShared.InvalidateExploreCache(ctx, r.cache)
	if r.cache != nil {
		_ = r.cache.InvalidateTags(ctx,
			infrarepoShared.CommunityMembersGlobalTag(),
			infrarepoShared.CommunityIDGlobalTag(),
		)
	}
	return nil
}

func (r *UserProfileRepository) FindByUserID(userID uint) (*userprofilemodel.UserProfile, error) {
	ctx := context.Background()
	key := userProfileKey(userID)
	var profile userprofilemodel.UserProfile
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &profile); err == nil {
			return &profile, nil
		}
	}
	if err := r.readDB.Preload("FacultyLocation").Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		r.cache.Set(ctx, key, profile, cache.DefaultTTL)
	}
	return &profile, nil
}

// UpdateFacultyLocation sets the user's home campus and invalidates the profile
// cache. Separate from Upsert because Upsert writes a fixed column list.
func (r *UserProfileRepository) UpdateFacultyLocation(userID uint, locationID *uint) error {
	if err := r.writeDB.Model(&userprofilemodel.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("faculty_location_id", locationID).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	return nil
}

func (r *UserProfileRepository) FindByUserIDs(ids []uint) ([]userprofilemodel.UserProfile, error) {
	if len(ids) == 0 {
		return []userprofilemodel.UserProfile{}, nil
	}
	var profiles []userprofilemodel.UserProfile
	err := r.readDB.Where("user_id IN ?", ids).Find(&profiles).Error
	return profiles, err
}

func (r *UserProfileRepository) ExistsByUserID(userID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&userprofilemodel.UserProfile{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func (r *UserProfileRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userProfileKey(userID))
}
