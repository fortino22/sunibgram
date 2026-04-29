package repository

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type LikeRepository struct {
	base *interactionRepo
}

func NewLikeRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.ILikeRepository {
	return &LikeRepository{base: newInteractionRepo(writeDB, readDB, c, "likes")}
}

func (r *LikeRepository) Save(like *model.Like) error {
	if err := r.base.saveRecord(like); err != nil {
		return err
	}
	r.base.invalidateUser(context.Background(), like.UserID)
	return nil
}

func (r *LikeRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &model.Like{}); err != nil {
		return err
	}
	r.base.invalidateUser(context.Background(), userID)
	return nil
}

func (r *LikeRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &model.Like{})
}

func (r *LikeRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
