package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
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

func (r *userFacultyRepository) Save(uf *usermodel.UserFaculty) error {
	if err := r.writeDB.Create(uf).Error; err != nil {
		return err
	}
	r.invalidateUser(uf.UserID)
	r.invalidateProfileMeta(uf.UserID)
	invalidateNotificationSenderCache(context.Background(), r.cache, uf.UserID)
	r.invalidateThreadCaches()
	return nil
}

func (r *userFacultyRepository) FindByUserID(userID uint) (*usermodel.UserFaculty, error) {
	ctx := context.Background()
	key := userFacultyKey(userID)
	var uf usermodel.UserFaculty
	if err := r.cache.Get(ctx, key, &uf); err == nil {
		return &uf, nil
	}
	if err := r.readDB.Preload("Faculty").Where("user_id = ?", userID).First(&uf).Error; err != nil {
		return nil, err
	}
	r.cache.Set(ctx, key, uf, cache.FacultyTTL)
	return &uf, nil
}

func (r *userFacultyRepository) FindByUserIDs(ids []uint) ([]usermodel.UserFaculty, error) {
	if len(ids) == 0 {
		return []usermodel.UserFaculty{}, nil
	}
	var rows []usermodel.UserFaculty
	err := r.readDB.Preload("Faculty").Where("user_id IN ?", ids).Find(&rows).Error
	return rows, err
}

func (r *userFacultyRepository) FindByFacultyID(facultyID uint, page, limit int) ([]usermodel.UserFaculty, int64, error) {
	var rows []usermodel.UserFaculty
	var total int64
	r.readDB.Model(&usermodel.UserFaculty{}).Where("faculty_id = ?", facultyID).Count(&total)
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
	result := r.writeDB.Delete(&usermodel.UserFaculty{}, "user_id = ? AND faculty_id = ?", userID, facultyID)
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
	_ = r.cache.Delete(ctx, "feed:trending:global", "reports:all")
	_ = r.cache.InvalidateTags(ctx,
		threadRootsGlobalTag(),
		threadCommunityGlobalTag(),
		threadAllCommunitiesGlobalTag(),
		threadMyCommunitiesGlobalTag(),
		threadFollowingGlobalTag(),
		communityThreadsGlobalTag(),
		eventThreadsGlobalTag(),
	)
	// thread:userfeed/usermedia:*, feed:foryou:*, reports:id:* still SCAN —
	// rare faculty assignment path, no global tag for those families.
	_ = r.cache.DeleteByPatterns(ctx,
		"thread:userfeed:*",
		"thread:usermedia:*",
		"feed:foryou:*",
		"reports:id:*",
	)
}
