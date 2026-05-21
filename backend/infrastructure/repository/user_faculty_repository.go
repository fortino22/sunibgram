package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type userFacultyRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewUserFacultyRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IUserFacultyRepository {
	return &userFacultyRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func userFacultyKey(userID uint) string {
	return fmt.Sprintf("userfaculty:user:%d", userID)
}

func (r *userFacultyRepository) Save(uf *model.UserFaculty) error {
	if err := r.writeDB.Create(uf).Error; err != nil {
		return err
	}
	r.invalidateUser(uf.UserID)
	r.invalidateProfileMeta(uf.UserID)
	invalidateNotificationSenderCache(context.Background(), r.cache, uf.UserID)
	r.invalidateThreadCaches()
	return nil
}

func (r *userFacultyRepository) FindByUserID(userID uint) (*model.UserFaculty, error) {
	ctx := context.Background()
	key := userFacultyKey(userID)
	var uf model.UserFaculty
	if err := r.cache.Get(ctx, key, &uf); err == nil {
		return &uf, nil
	}
	if err := r.readDB.Preload("Faculty").Where("user_id = ?", userID).First(&uf).Error; err != nil {
		return nil, err
	}
	r.cache.Set(ctx, key, uf, cache.FacultyTTL)
	return &uf, nil
}

func (r *userFacultyRepository) FindByUserIDs(ids []uint) ([]model.UserFaculty, error) {
	if len(ids) == 0 {
		return []model.UserFaculty{}, nil
	}
	var rows []model.UserFaculty
	err := r.readDB.Preload("Faculty").Where("user_id IN ?", ids).Find(&rows).Error
	return rows, err
}

func (r *userFacultyRepository) FindByFacultyID(facultyID uint, page, limit int) ([]model.UserFaculty, int64, error) {
	var rows []model.UserFaculty
	var total int64
	r.readDB.Model(&model.UserFaculty{}).Where("faculty_id = ?", facultyID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Where("faculty_id = ?", facultyID).
		Offset(offset).Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *userFacultyRepository) Delete(userID, facultyID uint) error {
	result := r.writeDB.Delete(&model.UserFaculty{}, "user_id = ? AND faculty_id = ?", userID, facultyID)
	if result.RowsAffected == 0 {
		return errors.New("assignment not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateUser(userID)
	r.invalidateProfileMeta(userID)
	invalidateNotificationSenderCache(context.Background(), r.cache, userID)
	r.invalidateThreadCaches()
	return nil
}

func (r *userFacultyRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userFacultyKey(userID))
}

func (r *userFacultyRepository) invalidateProfileMeta(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), profileMetaKey(userID))
}

func (r *userFacultyRepository) invalidateThreadCaches() {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPattern(ctx, "event:threads:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:following:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:roots:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:community:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:all-communities:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:my-communities:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:userfeed:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:usermedia:*")
	_ = r.cache.DeleteByPattern(ctx, "community:threads:top:*")
	_ = r.cache.DeleteByPattern(ctx, "community:threads:latest:*")
	_ = r.cache.DeleteByPattern(ctx, "feed:foryou:*")
	_ = r.cache.Delete(ctx, "feed:trending:global")
}
