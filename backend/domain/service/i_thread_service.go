package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IThreadService interface {
	Create(ctx context.Context, userID uint, content string, parentID *uint) (*model.Thread, error)
	GetByID(id uint) (*model.Thread, error)
	GetRoots(page, limit int) ([]model.Thread, int64, error)
	GetChain(rootID uint, page, limit int) ([]model.Thread, int64, error)
	Update(ctx context.Context, id, userID uint, content string) (*model.Thread, error)
	Delete(ctx context.Context, id, userID uint) error
	Repost(ctx context.Context, userID, threadID uint) error
	Unrepost(ctx context.Context, userID, threadID uint) error
	GetMyPosts(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
	GetMyReplies(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error)
}
