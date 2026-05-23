package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/infrastructure/cache"
)


type RateLimiter struct {
	cache  *cache.Cache
	prefix string
	max    int
	window time.Duration
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	switch {
	case m > 0 && s > 0:
		return fmt.Sprintf("%d minutes %d seconds", m, s)
	case m > 0:
		return fmt.Sprintf("%d minutes", m)
	default:
		return fmt.Sprintf("%d seconds", s)
	}
}

// New constructs a Redis-backed RateLimiter. `prefix` namespaces keys so
// different call sites do not collide (e.g. "otp_send", "otp_verify",
// "otp_fail"). `max` requests per `window` per key.
func New(c *cache.Cache, prefix string, max int, window time.Duration) *RateLimiter {
	return &RateLimiter{cache: c, prefix: prefix, max: max, window: window}
}

func (r *RateLimiter) redisKey(key string) string {
	return fmt.Sprintf("ratelimit:%s:%s", r.prefix, key)
}

// Allow consumes one slot and returns an error if the post-increment count
// exceeds the limit. Fail-open on Redis errors — better a brief lapse in rate
// limiting than a full auth outage.
func (r *RateLimiter) Allow(key string) error {
	ctx := context.Background()
	count, ttl, err := r.cache.IncrWithWindow(ctx, r.redisKey(key), r.window)
	if err != nil {
		return nil
	}
	if count > int64(r.max) {
		wait := ttl.Round(time.Second)
		if wait <= 0 {
			wait = r.window
		}
		return fmt.Errorf("try again in %s", formatDuration(wait))
	}
	return nil
}

// Check reports whether the key is at or over the limit without consuming a
// slot. Returns nil when the counter is missing or under the cap.
func (r *RateLimiter) Check(key string) error {
	ctx := context.Background()
	count, ttl, err := r.cache.CounterTTL(ctx, r.redisKey(key))
	if err != nil {
		return nil
	}
	if count >= int64(r.max) {
		wait := ttl.Round(time.Second)
		if wait <= 0 {
			wait = r.window
		}
		return fmt.Errorf("try again in %s", formatDuration(wait))
	}
	return nil
}

// Fail increments the counter without gating; pair with Check to record a
// failed attempt against a key that already passed an entry check.
func (r *RateLimiter) Fail(key string) {
	ctx := context.Background()
	_, _, _ = r.cache.IncrWithWindow(ctx, r.redisKey(key), r.window)
}
