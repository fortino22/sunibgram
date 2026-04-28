package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
)

type BookmarkRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewBookmarkRepository(writeDB, readDB *gorm.DB) domainrepo.IBookmarkRepository {
	return &BookmarkRepository{writeDB: writeDB, readDB: readDB}
}

func (r *BookmarkRepository) Save(bookmark *model.Bookmark) error {
	return r.writeDB.Create(bookmark).Error
}

func (r *BookmarkRepository) Delete(userID, threadID uint) error {
	return r.writeDB.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&model.Bookmark{}).Error
}

func (r *BookmarkRepository) Exists(userID, threadID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.Bookmark{}).Where("user_id = ? AND thread_id = ?", userID, threadID).Count(&count).Error
	return count > 0, err
}

func (r *BookmarkRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return findThreadsByInteraction(r.readDB, "bookmarks", userID, page, limit)
}
