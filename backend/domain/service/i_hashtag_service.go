package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IHashtagService interface {
	GetTrending(ctx context.Context, limit int) ([]model.TrendingHashtag, error)
	GetThreadsByHashtag(ctx context.Context, name string, cursor *uint, userID uint, limit int) ([]model.Thread, *uint, error)
}
