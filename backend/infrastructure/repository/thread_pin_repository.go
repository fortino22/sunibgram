package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type ThreadPinRepository struct {
	writeDB  *gorm.DB
	readDB   *gorm.DB
	threadDB *threadDB
}

func NewThreadPinRepository(writeDB, readDB *gorm.DB) domainrepo.IThreadPinRepository {
	return &ThreadPinRepository{
		writeDB:  writeDB,
		readDB:   readDB,
		threadDB: &threadDB{writeDB: writeDB, readDB: readDB},
	}
}

func (r *ThreadPinRepository) Pin(pin *model.ThreadPin) error {
	return r.writeDB.Create(pin).Error
}

func (r *ThreadPinRepository) Unpin(userID, threadID uint) error {
	return r.writeDB.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&model.ThreadPin{}).Error
}

func (r *ThreadPinRepository) ExistsPin(userID, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.ThreadPin{}).
		Where("user_id = ? AND thread_id = ?", userID, threadID).
		Count(&count).Error
	return count > 0, err
}

func (r *ThreadPinRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	var total int64
	r.readDB.Model(&model.ThreadPin{}).Where("user_id = ?", userID).Count(&total)

	var pins []model.ThreadPin
	offset := (page - 1) * limit
	if err := r.readDB.
		Preload("Thread.User").
		Preload("Thread.OriginalThread.User").
		Preload("Thread.Community").
		Preload("Thread.OriginalThread.Community").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&pins).Error; err != nil {
		return nil, nil, 0, err
	}
	if len(pins) == 0 {
		return nil, nil, total, nil
	}

	threads := make([]model.Thread, len(pins))
	timestamps := make([]time.Time, len(pins))
	for i, pin := range pins {
		threads[i] = pin.Thread
		timestamps[i] = pin.CreatedAt
	}
	r.threadDB.attachFaculties(threads)

	return threads, timestamps, total, nil
}
