package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserProfileRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewUserProfileRepository(writeDB *gorm.DB, readDB *gorm.DB) domainrepo.IUserProfileRepository {
	return &UserProfileRepository{writeDB: writeDB, readDB: readDB}
}

func (r *UserProfileRepository) Upsert(profile *model.UserProfile) error {
	if profile.ProfileID != 0 {
		return r.writeDB.Model(profile).Updates(map[string]any{
			"avatar_object_name": profile.AvatarObjectName,
			"avatar_bucket":      profile.AvatarBucket,
			"banner_object_name": profile.BannerObjectName,
			"banner_bucket":      profile.BannerBucket,
			"bio":                profile.Bio,
		}).Error
	}
	return r.writeDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"avatar_object_name", "avatar_bucket", "banner_object_name", "banner_bucket", "bio", "updated_at"}),
	}).Create(profile).Error
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
