package facultylocationrepo

import (
	"context"

	"gorm.io/gorm"

	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
	facultylocationrepoiface "github.com/sunibgram/domain/repository/faculty_location"
	"github.com/sunibgram/infrastructure/cache"
)

type facultyLocationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewFacultyLocationRepository(writeDB, readDB *gorm.DB, c ...*cache.Cache) facultylocationrepoiface.IFacultyLocationRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &facultyLocationRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

const (
	facultyLocationsSelectableKey = "faculty_locations:list:selectable"
	facultyLocationsAudienceKey   = "faculty_locations:list:audience"
)

const facultyLocationOrder = "sort_order ASC, region ASC, name ASC"

func (r *facultyLocationRepository) FindAll(includeAll bool) ([]facultylocationmodel.FacultyLocation, error) {
	ctx := context.Background()
	key := facultyLocationsSelectableKey
	if includeAll {
		key = facultyLocationsAudienceKey
	}
	var locations []facultylocationmodel.FacultyLocation
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &locations); err == nil {
			return locations, nil
		}
	}
	q := r.readDB.Order(facultyLocationOrder)
	if !includeAll {
		q = q.Where("is_all_regions = ?", false)
	}
	if err := q.Find(&locations).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, locations, cache.FacultyLocationTTL)
	}
	return locations, nil
}

func (r *facultyLocationRepository) FindByIDs(ids []uint) ([]facultylocationmodel.FacultyLocation, error) {
	if len(ids) == 0 {
		return []facultylocationmodel.FacultyLocation{}, nil
	}
	var locations []facultylocationmodel.FacultyLocation
	err := r.readDB.Where("id IN ?", ids).Order(facultyLocationOrder).Find(&locations).Error
	return locations, err
}

func (r *facultyLocationRepository) ExistsSelectable(id uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&facultylocationmodel.FacultyLocation{}).
		Where("id = ? AND is_all_regions = ?", id, false).Count(&count).Error
	return count > 0, err
}

func (r *facultyLocationRepository) Exists(id uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&facultylocationmodel.FacultyLocation{}).
		Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// InvalidateFacultyLocations drops both cached list keys. Called from an admin
// mutate path when/if faculty-location entities are edited.
func (r *facultyLocationRepository) InvalidateFacultyLocations(ctx context.Context) {
	if r.cache == nil {
		return
	}
	_ = r.cache.Delete(ctx, facultyLocationsSelectableKey, facultyLocationsAudienceKey)
}
