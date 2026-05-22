package service

import (
	"context"

	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
)

type IHashtagService interface {
	GetTrending(ctx context.Context, limit int) ([]hashtagmodel.TrendingHashtag, error)
	InvalidateTrendingCache(ctx context.Context)
}
