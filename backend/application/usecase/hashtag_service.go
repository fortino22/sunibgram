package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/infrastructure/cache"
)

const (
	defaultHashtagLimit = 20
	trendingHashtagTTL  = 5 * time.Minute
)

type HashtagService struct {
	hashtagRepo repository.IHashtagRepository
	cache       *cache.Cache
}

func NewHashtagService(hashtagRepo repository.IHashtagRepository, c *cache.Cache) service.IHashtagService {
	return &HashtagService{hashtagRepo: hashtagRepo, cache: c}
}

func (s *HashtagService) GetTrending(ctx context.Context, limit int) ([]model.TrendingHashtag, error) {
	_, limit = helper.NormalizePageLimit(1, limit, defaultHashtagLimit)

	key := fmt.Sprintf("trending_hashtags:%d", limit)
	var cached []model.TrendingHashtag
	if s.cache != nil {
		if err := s.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}

	now := time.Now()
	hashtags, err := s.hashtagRepo.GetTrending(ctx, limit, now.Add(-24*time.Hour), now.Add(-7*24*time.Hour))
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, key, hashtags, trendingHashtagTTL)
	}
	return hashtags, nil
}

func (s *HashtagService) InvalidateTrendingCache(ctx context.Context) {
	if s.cache != nil {
		s.cache.DeleteByPattern(ctx, "trending_hashtags:*")
	}
}
