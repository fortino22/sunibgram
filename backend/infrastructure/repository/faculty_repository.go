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

type facultyRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

type cachedFaculties struct {
	Faculties []model.Faculty
	Total     int64
}

func NewFacultyRepository(writeDB, readDB *gorm.DB, c ...*cache.Cache) domainrepo.IFacultyRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &facultyRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

func facultyIDKey(id uint) string {
	return fmt.Sprintf("faculty:id:%d", id)
}

func facultyListKey(page, limit int) string {
	return fmt.Sprintf("faculty:list:page:%d:limit:%d", page, limit)
}

func (r *facultyRepository) Save(faculty *model.Faculty) error {
	if err := r.writeDB.Create(faculty).Error; err != nil {
		return err
	}
	r.invalidateFacultyList()
	return nil
}

func (r *facultyRepository) FindAll(page, limit int) ([]model.Faculty, int64, error) {
	ctx := context.Background()
	key := facultyListKey(page, limit)
	var cached cachedFaculties
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached.Faculties, cached.Total, nil
		}
	}

	var faculties []model.Faculty
	var total int64
	r.readDB.Model(&model.Faculty{}).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Offset(offset).Limit(limit).Find(&faculties).Error; err != nil {
		return nil, 0, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, cachedFaculties{Faculties: faculties, Total: total}, cache.FacultyTTL)
	}
	return faculties, total, nil
}

func (r *facultyRepository) FindByID(id uint) (*model.Faculty, error) {
	ctx := context.Background()
	key := facultyIDKey(id)
	var faculty model.Faculty
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &faculty); err == nil {
			return &faculty, nil
		}
	}
	if err := r.readDB.First(&faculty, "faculty_id = ?", id).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, faculty, cache.FacultyTTL)
	}
	return &faculty, nil
}

func (r *facultyRepository) Update(faculty *model.Faculty) error {
	result := r.writeDB.Model(faculty).Where("faculty_id = ?", faculty.FacultyID).
		UpdateColumn("faculty_name", faculty.FacultyName)
	if result.RowsAffected == 0 {
		return errors.New("faculty not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateFacultyUpdate(faculty.FacultyID)
	return nil
}

func (r *facultyRepository) Delete(id uint) error {
	result := r.writeDB.Delete(&model.Faculty{}, "faculty_id = ?", id)
	if result.RowsAffected == 0 {
		return errors.New("faculty not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateFacultyUpdate(id)
	return nil
}

func (r *facultyRepository) UpdateTotalUsers(facultyID uint, delta int) error {
	var expr interface{}
	if delta >= 0 {
		expr = gorm.Expr("total_users + ?", delta)
	} else {
		expr = gorm.Expr("GREATEST(total_users + ?, 0)", delta)
	}
	if err := r.writeDB.Model(&model.Faculty{}).
		Where("faculty_id = ?", facultyID).
		UpdateColumn("total_users", expr).Error; err != nil {
		return err
	}
	r.invalidateFacultyUpdate(facultyID)
	return nil
}

func (r *facultyRepository) invalidateFacultyList() {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPattern(ctx, "faculty:list:*")
	invalidateExploreCache(ctx, r.cache)
}

func (r *facultyRepository) invalidateFacultyUpdate(id uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, facultyIDKey(id))
	_ = r.cache.DeleteByPattern(ctx, "faculty:list:*")
	_ = r.cache.DeleteByPattern(ctx, "profile:user:*")
	_ = r.cache.DeleteByPattern(ctx, "userfaculty:user:*")
	invalidateNotificationSenderCachePattern(ctx, r.cache)
	invalidateExploreCache(ctx, r.cache)
}
