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
	DeleteByPatterns(ctx context.Context, patterns ...string) error

	// Reverse-index tag invalidation (phase 2). RegisterCacheKey associates a
	// cache key with one or more tags on Set. InvalidateTag(s) drops every
	// member key of a tag in O(set size) — no SCAN required. Tag names follow
	// the "tag:<entity>:<id>[:scope]" convention (see infrastructure/repository
	// tag builders). Tag TTL is bounded; longest cache TTL must not exceed it.
	RegisterCacheKey(ctx context.Context, tag, cacheKey string) error
	RegisterCacheKeyMany(ctx context.Context, tags []string, cacheKey string) error
	InvalidateTag(ctx context.Context, tag string) error
	InvalidateTags(ctx context.Context, tags ...string) error
}
