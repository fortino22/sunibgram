package hashtaguc

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/application/helper"
	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	hashtagrepoiface "github.com/sunibgram/domain/repository/hashtag"
	hashtagsvciface "github.com/sunibgram/domain/service/hashtag"
	"github.com/sunibgram/infrastructure/cache"
)

const (
	trendingHashtagLimit = 30
	trendingHashtagTTL   = 5 * time.Minute
)

type HashtagService struct {
	hashtagRepo hashtagrepoiface.IHashtagRepository
	cache       *cache.Cache
}

func NewHashtagService(hashtagRepo hashtagrepoiface.IHashtagRepository, c *cache.Cache) hashtagsvciface.IHashtagService {
	return &HashtagService{hashtagRepo: hashtagRepo, cache: c}
}

func (s *HashtagService) GetTrending(ctx context.Context, limit int) ([]hashtagmodel.TrendingHashtag, error) {
	_, limit = helper.NormalizePageLimit(1, limit, trendingHashtagLimit)
	if limit > trendingHashtagLimit {
		limit = trendingHashtagLimit
	}

	key := fmt.Sprintf("trending_hashtags:%d", limit)
	var cached []hashtagmodel.TrendingHashtag
	if s.cache != nil {
		if err := s.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}

	hashtags, err := s.hashtagRepo.GetTrending(ctx, limit)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if err := s.cache.Set(ctx, key, hashtags, trendingHashtagTTL); err == nil {
			// Tag mirror of trendingHashtagsTag() in
			// infrastructure/repository/tag_keys.go. Keep these strings in sync.
			_ = s.cache.RegisterCacheKey(ctx, "tag:trending-hashtags", key)
		}
	}
	return hashtags, nil
}

func (s *HashtagService) InvalidateTrendingCache(ctx context.Context) {
	if s.cache != nil {
		_ = s.cache.InvalidateTag(ctx, "tag:trending-hashtags")
	}
}
