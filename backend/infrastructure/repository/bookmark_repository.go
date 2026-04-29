package repository

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type BookmarkRepository struct {
	base *interactionRepo
}

func NewBookmarkRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IBookmarkRepository {
	return &BookmarkRepository{base: newInteractionRepo(writeDB, readDB, c, "bookmarks")}
}

func (r *BookmarkRepository) Save(bookmark *model.Bookmark) error {
	if err := r.base.saveRecord(bookmark); err != nil {
		return err
	}
	r.base.invalidateUser(context.Background(), bookmark.UserID)
	return nil
}

func (r *BookmarkRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &model.Bookmark{}); err != nil {
		return err
	}
	r.base.invalidateUser(context.Background(), userID)
	return nil
}

func (r *BookmarkRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &model.Bookmark{})
}

func (r *BookmarkRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
