package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type LikeRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewLikeRepository(writeDB, readDB *gorm.DB) domainrepo.ILikeRepository {
	return &LikeRepository{writeDB: writeDB, readDB: readDB}
}

func (r *LikeRepository) Save(like *model.Like) error {
	return r.writeDB.Create(like).Error
}

func (r *LikeRepository) Delete(userID, threadID uint) error {
	return r.writeDB.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&model.Like{}).Error
}

func (r *LikeRepository) Exists(userID, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.Like{}).Where("user_id = ? AND thread_id = ?", userID, threadID).Count(&count).Error
	return count > 0, err
}

func (r *LikeRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return findThreadsByInteraction(r.readDB, "likes", userID, page, limit)
}
