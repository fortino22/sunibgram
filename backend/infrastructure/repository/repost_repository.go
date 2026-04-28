package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type RepostRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewRepostRepository(writeDB, readDB *gorm.DB) domainrepo.IRepostRepository {
	return &RepostRepository{writeDB: writeDB, readDB: readDB}
}

func (r *RepostRepository) Save(repost *model.Repost) error {
	return r.writeDB.Create(repost).Error
}

func (r *RepostRepository) Delete(userID, threadID uint) error {
	return r.writeDB.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&model.Repost{}).Error
}

func (r *RepostRepository) Exists(userID, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.Repost{}).Where("user_id = ? AND thread_id = ?", userID, threadID).Count(&count).Error
	return count > 0, err
}

func (r *RepostRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return findThreadsByInteraction(r.readDB, "reposts", userID, page, limit)
}
