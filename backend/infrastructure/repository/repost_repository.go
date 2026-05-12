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
	base       *interactionRepo
	patchCount func(threadID uint, column string, delta int)
}

func NewRepostRepository(writeDB, readDB *gorm.DB, c *cache.Cache, patchCount func(uint, string, int)) domainrepo.IRepostRepository {
	return &RepostRepository{
		base:       newInteractionRepo(writeDB, readDB, c, "reposts", &threadDB{writeDB: writeDB, readDB: readDB}),
		patchCount: patchCount,
	}
}

func (r *RepostRepository) Save(repost *model.Repost) error {
	if err := r.base.saveRecord(repost); err != nil {
		return err
	}
	r.base.writeDB.Model(&model.Thread{}).Where("thread_id = ?", repost.ThreadID).
		UpdateColumn("repost_count", gorm.Expr("repost_count + 1"))
	patchEventCounterByThreadID(r.base.writeDB, r.base.cache, repost.ThreadID, "repost_count", 1)
	patchHashtagCounterByThreadID(r.base.writeDB, r.base.cache, repost.ThreadID, "repost_count", 1)
	ctx := context.Background()
	r.base.invalidateUser(ctx, repost.UserID)
	r.patchCount(repost.ThreadID, "repost_count", 1)
	return nil
}

func (r *RepostRepository) Delete(userID, threadID uint) error {
	if err := r.base.deleteRecord(userID, threadID, &model.Repost{}); err != nil {
		return err
	}
	r.base.writeDB.Model(&model.Thread{}).Where("thread_id = ?", threadID).
		UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
	patchEventCounterByThreadID(r.base.writeDB, r.base.cache, threadID, "repost_count", -1)
	patchHashtagCounterByThreadID(r.base.writeDB, r.base.cache, threadID, "repost_count", -1)
	ctx := context.Background()
	r.base.invalidateUser(ctx, userID)
	r.patchCount(threadID, "repost_count", -1)
	return nil
}

func (r *RepostRepository) Exists(userID, threadID uint) (bool, error) {
	return r.base.exists(userID, threadID, &model.Repost{})
}

func (r *RepostRepository) FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error) {
	return r.base.findByUserID(userID, page, limit)
}
