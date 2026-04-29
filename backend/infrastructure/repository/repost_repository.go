package repository

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type RepostRepository struct {
	base *interactionRepo
}

func NewRepostRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IRepostRepository {
	return &RepostRepository{base: newInteractionRepo(writeDB, readDB, c, "reposts")}
}

func (r *RepostRepository) Save(repost *model.Repost) error {
	if err := r.base.saveRecord(repost); err != nil {
		return err
	}
	r.base.invalidateUser(context.Background(), repost.UserID)
	return nil
}

func (r *RepostRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &model.Repost{}); err != nil {
		return err
	}
	r.base.invalidateUser(context.Background(), userID)
	return nil
}

func (r *RepostRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &model.Repost{})
}

func (r *RepostRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
