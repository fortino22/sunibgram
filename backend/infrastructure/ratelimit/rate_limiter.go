package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

type bucket struct {
	count     int
	windowEnd time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	max     int
	window  time.Duration
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

func New(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		max:     max,
		window:  window,
	}
}

// Allow increments the counter and returns an error if the limit is reached.
func (r *RateLimiter) Allow(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[key]
	if !ok || now.After(b.windowEnd) {
		r.buckets[key] = &bucket{count: 1, windowEnd: now.Add(r.window)}
		return nil
	}

	if b.count >= r.max {
		wait := time.Until(b.windowEnd).Round(time.Second)
		return fmt.Errorf("try again in %s", formatDuration(wait))
	}

	b.count++
	return nil
}

// Check returns an error if the key is already at or over the limit, without incrementing.
func (r *RateLimiter) Check(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[key]
	if !ok || now.After(b.windowEnd) {
		return nil
	}
	if b.count >= r.max {
		wait := time.Until(b.windowEnd).Round(time.Second)
		return fmt.Errorf("try again in %s", formatDuration(wait))
	}
	return nil
}

// Fail increments the failure counter for the key without gating on the limit.
// Use after a failed attempt to record it; pair with Check to gate entry.
func (r *RateLimiter) Fail(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[key]
	if !ok || now.After(b.windowEnd) {
		r.buckets[key] = &bucket{count: 1, windowEnd: now.Add(r.window)}
		return
	}
	b.count++
}
