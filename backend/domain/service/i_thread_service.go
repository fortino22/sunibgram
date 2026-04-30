package service

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
)

type IThreadService interface {
	Create(ctx context.Context, userID uint, content string, parentID *uint) (*model.Thread, error)
	GetByID(id uint) (*model.Thread, error)
	GetRoots(page, limit int) ([]model.Thread, int64, error)
	GetChain(rootID uint, page, limit int) ([]model.Thread, int64, error)
	Update(ctx context.Context, id, userID uint, content string) (*model.Thread, error)
	Delete(ctx context.Context, id, userID uint) error
	GetMyPosts(userID uint, page, limit int) ([]model.Thread, []*time.Time, int64, error)
	GetMyReplies(userID uint, page, limit int) (threads []model.Thread, repostTimestamps []*time.Time, parents []*model.Thread, total int64, err error)
}
