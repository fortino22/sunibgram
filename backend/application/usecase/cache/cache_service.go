package cacheuc

import (
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	cachesvciface "github.com/sunibgram/domain/service/cache"
)

type CacheService struct {
	cacheRepo cacherepoiface.ICacheRepository
}

func NewCacheService(cacheRepo cacherepoiface.ICacheRepository) cachesvciface.ICacheService {
	return &CacheService{cacheRepo: cacheRepo}
}

func (s *CacheService) ClearAll() error {
	return s.cacheRepo.ClearAll()
}
