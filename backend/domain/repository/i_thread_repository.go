package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type IThreadRepository interface {
	Save(thread *model.Thread) error
	FindByID(id uint) (*model.Thread, error)
	FindRoots(page, limit int) ([]model.Thread, int64, error)
	FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error)
	Update(thread *model.Thread) error
	Delete(id uint) error
	FindPostsByUserID(userID uint, page, limit int) ([]model.Thread, []*time.Time, int64, error)
	FindReplyTabByUserID(userID uint, page, limit int) (threads []model.Thread, repostTimestamps []*time.Time, parents []*model.Thread, total int64, err error)
}
