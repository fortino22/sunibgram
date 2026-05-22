package repository

import (
	"context"
	"time"

	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
)

type CacheRepository struct {
	cache *cache.Cache
}

func NewCacheRepository(c *cache.Cache) domainrepo.ICacheRepository {
	return &CacheRepository{cache: c}
}

func (r *CacheRepository) ClearAll() error {
	return r.cache.FlushDB(context.Background())
}

func (r *CacheRepository) Get(ctx context.Context, key string, dest any) error {
	return r.cache.Get(ctx, key, dest)
}

func (r *CacheRepository) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.cache.Set(ctx, key, value, ttl)
}

func (r *CacheRepository) Delete(ctx context.Context, keys ...string) error {
	return r.cache.Delete(ctx, keys...)
}

func (r *CacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	return r.cache.DeleteByPattern(ctx, pattern)
}

func (r *CacheRepository) DeleteByPatterns(ctx context.Context, patterns ...string) error {
	return r.cache.DeleteByPatterns(ctx, patterns...)
}

func (r *CacheRepository) RegisterCacheKey(ctx context.Context, tag, cacheKey string) error {
	return r.cache.RegisterCacheKey(ctx, tag, cacheKey)
}

func (r *CacheRepository) RegisterCacheKeyMany(ctx context.Context, tags []string, cacheKey string) error {
	return r.cache.RegisterCacheKeyMany(ctx, tags, cacheKey)
}

func (r *CacheRepository) InvalidateTag(ctx context.Context, tag string) error {
	return r.cache.InvalidateTag(ctx, tag)
}

func (r *CacheRepository) InvalidateTags(ctx context.Context, tags ...string) error {
	return r.cache.InvalidateTags(ctx, tags...)
}
