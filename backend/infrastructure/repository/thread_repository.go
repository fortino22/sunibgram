package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type ThreadRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewThreadRepository(writeDB, readDB *gorm.DB) domainrepo.IThreadRepository {
	return &ThreadRepository{writeDB: writeDB, readDB: readDB}
}

func (r *ThreadRepository) Save(thread *model.Thread) error {
	return r.writeDB.Create(thread).Error
}

func (r *ThreadRepository) FindByID(id uint) (*model.Thread, error) {
	var thread model.Thread
	err := r.readDB.Preload("User").First(&thread, id).Error
	return &thread, err
}

// FindRoots returns only head posts (parent_id IS NULL), newest first.
func (r *ThreadRepository) FindRoots(page, limit int) ([]model.Thread, int64, error) {
	var threads []model.Thread
	var total int64

	r.readDB.Model(&model.Thread{}).Where("parent_id IS NULL").Count(&total)

	offset := (page - 1) * limit
	err := r.readDB.Preload("User").
		Where("parent_id IS NULL").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error

	return threads, total, err
}

// FindByRootID returns all posts in a thread chain ordered chronologically, paginated.
func (r *ThreadRepository) FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error) {
	var threads []model.Thread
	var total int64

	base := r.readDB.Model(&model.Thread{}).Where("thread_id = ? OR root_id = ?", rootID, rootID)
	base.Count(&total)

	offset := (page - 1) * limit
	err := r.readDB.Preload("User").
		Where("thread_id = ? OR root_id = ?", rootID, rootID).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Find(&threads).Error

	return threads, total, err
}

func (r *ThreadRepository) Update(thread *model.Thread) error {
	return r.writeDB.Save(thread).Error
}

func (r *ThreadRepository) Delete(id uint) error {
	return r.writeDB.Delete(&model.Thread{}, id).Error
}
