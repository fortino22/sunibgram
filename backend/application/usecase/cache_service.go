package usecase

import (
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type CacheService struct {
	cacheRepo repository.ICacheRepository
}

func NewCacheService(cacheRepo repository.ICacheRepository) service.ICacheService {
	return &CacheService{cacheRepo: cacheRepo}
}

func (s *CacheService) ClearAll() error {
	return s.cacheRepo.ClearAll()
}
