package repository

import (
	"context"

	"gorm.io/gorm"

	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
)

type ThreadMediaRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewThreadMediaRepository(writeDB, readDB *gorm.DB, c ...*cache.Cache) domainrepo.IThreadMediaRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &ThreadMediaRepository{writeDB: writeDB, readDB: readDB, cache: redisCache}
}

func (r *ThreadMediaRepository) SaveAll(media []threadmediamodel.ThreadMedia) error {
	if err := r.writeDB.Create(&media).Error; err != nil {
		return err
	}
	invalidateExploreCache(context.Background(), r.cache)
	return nil
}

func (r *ThreadMediaRepository) FindByThreadID(threadID uint) ([]threadmediamodel.ThreadMedia, error) {
	var media []threadmediamodel.ThreadMedia
	err := r.readDB.Where("thread_id = ?", threadID).Find(&media).Error
	return media, err
}

func (r *ThreadMediaRepository) FindByThreadIDs(threadIDs []uint) ([]threadmediamodel.ThreadMedia, error) {
	var media []threadmediamodel.ThreadMedia
	if len(threadIDs) == 0 {
		return media, nil
	}
	err := r.readDB.Where("thread_id IN ?", threadIDs).Find(&media).Error
	return media, err
}

func (r *ThreadMediaRepository) DeleteByThreadID(threadID uint) error {
	if err := r.writeDB.Where("thread_id = ?", threadID).Delete(&threadmediamodel.ThreadMedia{}).Error; err != nil {
		return err
	}
	invalidateExploreCache(context.Background(), r.cache)
	return nil
}

func (r *ThreadMediaRepository) DeleteByIDs(ids []uint) ([]threadmediamodel.ThreadMedia, error) {
	var media []threadmediamodel.ThreadMedia
	if err := r.readDB.Where("media_id IN ?", ids).Find(&media).Error; err != nil {
		return nil, err
	}
	if err := r.writeDB.Where("media_id IN ?", ids).Delete(&threadmediamodel.ThreadMedia{}).Error; err != nil {
		return nil, err
	}
	invalidateExploreCache(context.Background(), r.cache)
	return media, nil
}
