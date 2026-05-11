package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IHashtagService interface {
	GetTrending(ctx context.Context, limit int) ([]model.TrendingHashtag, error)
	InvalidateTrendingCache(ctx context.Context)
}
