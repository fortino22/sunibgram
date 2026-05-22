package repository

import (
	"errors"

	"gorm.io/gorm"

	threadpinmodel "github.com/sunibgram/domain/model/thread_pin"
	threadpinrepoiface "github.com/sunibgram/domain/repository/thread_pin"
	"github.com/sunibgram/infrastructure/cache"
)

type ThreadPinRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewThreadPinRepository(writeDB, readDB *gorm.DB, c ...*cache.Cache) threadpinrepoiface.IThreadPinRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &ThreadPinRepository{
		writeDB: writeDB,
		readDB:  readDB,
		cache:   redisCache,
	}
}

func (r *ThreadPinRepository) Pin(pin *threadpinmodel.ThreadPin) error {
	if err := r.writeDB.Create(pin).Error; err != nil {
		return err
	}
	ViewerFlagAdd(r.cache, pin.UserID, "pinned", pin.ThreadID)
	return nil
}

func (r *ThreadPinRepository) Unpin(userID, threadID uint) error {
	if err := r.writeDB.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&threadpinmodel.ThreadPin{}).Error; err != nil {
		return err
	}
	ViewerFlagRemove(r.cache, userID, "pinned", threadID)
	return nil
}

func (r *ThreadPinRepository) ExistsPin(userID, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&threadpinmodel.ThreadPin{}).
		Where("user_id = ? AND thread_id = ?", userID, threadID).
		Count(&count).Error
	return count > 0, err
}

func (r *ThreadPinRepository) ArePinnedFor(userID uint, threadIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(threadIDs))
	if len(threadIDs) == 0 {
		return result, nil
	}

	var ids []uint
	err := r.readDB.Model(&threadpinmodel.ThreadPin{}).
		Select("thread_id").
		Where("user_id = ? AND thread_id IN ?", userID, threadIDs).
		Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

func (r *ThreadPinRepository) AllPinnedThreadIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.readDB.Model(&threadpinmodel.ThreadPin{}).
		Select("thread_id").
		Where("user_id = ?", userID).
		Scan(&ids).Error
	return ids, err
}

func (r *ThreadPinRepository) FindPinnedByUserID(userID uint) (*threadpinmodel.ThreadPin, error) {
	var pin threadpinmodel.ThreadPin
	err := r.readDB.Where("user_id = ?", userID).First(&pin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pin, nil
}
