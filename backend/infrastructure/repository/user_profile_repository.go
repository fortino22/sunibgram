package repository

import (
	"context"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserProfileRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewUserProfileRepository(writeDB *gorm.DB, readDB *gorm.DB, c ...*cache.Cache) domainrepo.IUserProfileRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &UserProfileRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

func (r *UserProfileRepository) Upsert(profile *model.UserProfile) error {
	if profile.ProfileID != 0 {
		if err := r.writeDB.Model(profile).Updates(map[string]any{
			"avatar_object_name": profile.AvatarObjectName,
			"avatar_bucket":      profile.AvatarBucket,
			"banner_object_name": profile.BannerObjectName,
			"banner_bucket":      profile.BannerBucket,
			"bio":                profile.Bio,
		}).Error; err != nil {
			return err
		}
		r.invalidateUser(profile.UserID)
		return nil
	}
	if err := r.writeDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"avatar_object_name", "avatar_bucket", "banner_object_name", "banner_bucket", "bio", "updated_at"}),
	}).Create(profile).Error; err != nil {
		return err
	}
	r.invalidateUser(profile.UserID)
	return nil
}

func (r *UserProfileRepository) FindByUserID(userID uint) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := r.readDB.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

func (r *UserProfileRepository) FindByUserIDs(ids []uint) ([]model.UserProfile, error) {
	if len(ids) == 0 {
		return []model.UserProfile{}, nil
	}
	var profiles []model.UserProfile
	err := r.readDB.Where("user_id IN ?", ids).Find(&profiles).Error
	return profiles, err
}

func (r *UserProfileRepository) ExistsByUserID(userID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.UserProfile{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func (r *UserProfileRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userKey(userID))
}
