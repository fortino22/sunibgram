package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type ThreadMediaRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewThreadMediaRepository(writeDB, readDB *gorm.DB) domainrepo.IThreadMediaRepository {
	return &ThreadMediaRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ThreadMediaRepository) SaveAll(media []model.ThreadMedia) error {
	return r.writeDB.Create(&media).Error
}

func (r *ThreadMediaRepository) UpdateOne(media *model.ThreadMedia) error {
	return r.writeDB.Save(media).Error
}

func (r *ThreadMediaRepository) FindByThreadID(threadID uint) ([]model.ThreadMedia, error) {
	var media []model.ThreadMedia
	err := r.readDB.Where("thread_id = ?", threadID).Find(&media).Error
	return media, err
}

func (r *ThreadMediaRepository) FindByThreadIDs(threadIDs []uint) ([]model.ThreadMedia, error) {
	var media []model.ThreadMedia
	if len(threadIDs) == 0 {
		return media, nil
	}
	err := r.readDB.Where("thread_id IN ?", threadIDs).Find(&media).Error
	return media, err
}

func (r *ThreadMediaRepository) DeleteByThreadID(threadID uint) error {
	return r.writeDB.Where("thread_id = ?", threadID).Delete(&model.ThreadMedia{}).Error
}

func (r *ThreadMediaRepository) DeleteByIDs(ids []uint) ([]model.ThreadMedia, error) {
	var media []model.ThreadMedia
	if err := r.readDB.Where("media_id IN ?", ids).Find(&media).Error; err != nil {
		return nil, err
	}
	if err := r.writeDB.Where("media_id IN ?", ids).Delete(&model.ThreadMedia{}).Error; err != nil {
		return nil, err
	}
	return media, nil
}
