package repository

import (
	"context"
	"time"
)

type ICacheRepository interface {
	ClearAll() error
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
}
