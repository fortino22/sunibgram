package repository

import (
	"context"

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
