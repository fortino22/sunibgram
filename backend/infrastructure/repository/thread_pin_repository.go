package repository

import (
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type ThreadPinRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewThreadPinRepository(writeDB, readDB *gorm.DB) domainrepo.IThreadPinRepository {
	return &ThreadPinRepository{
		writeDB: writeDB,
		readDB:  readDB,
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
