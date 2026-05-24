package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const DefaultTTL = 5 * time.Minute
const FacultyTTL = 24 * time.Hour

type Cache struct {
	client *redis.Client
}

func New() *Cache {
	addr := os.Getenv("REDIS_ADDR")

	password := os.Getenv("REDIS_PASSWORD")
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	b, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, b, ttl).Err()
}

// SetNX sets the key only if it does not already exist, with the given TTL.
// Returns true if the key was created (caller holds the lock), false if it
// already existed. Used for webhook idempotency and similar one-shot guards.
func (c *Cache) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return c.client.SetNX(ctx, key, b, ttl).Result()
}

func (c *Cache) RPush(ctx context.Context, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.RPush(ctx, key, b).Err()
}

func (c *Cache) BLPop(ctx context.Context, key string, timeout time.Duration) (string, error) {
	values, err := c.client.BLPop(ctx, timeout, key).Result()
	if err != nil {
		return "", err
	}
	if len(values) < 2 {
		return "", fmt.Errorf("cache blpop: empty result")
	}
	return values[1], nil
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Cache) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// DeleteByPattern removes all keys matching a glob pattern (e.g. "thread:roots:*").
func (c *Cache) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 1000).Result()
		if err != nil {
			return fmt.Errorf("cache scan: %w", err)
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("cache delete: %w", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// DeleteByPatterns fans SCAN across all patterns in parallel and pipelines the
// resulting DELs. Use when an invalidation path needs to drop multiple
// patterns at once — equivalent to N serial DeleteByPattern calls but much
// faster (parallel SCAN + single DEL pipeline per scan batch). Returns the
// first error encountered; other goroutines still complete.
func (c *Cache) DeleteByPatterns(ctx context.Context, patterns ...string) error {
	if len(patterns) == 0 {
		return nil
	}
	if len(patterns) == 1 {
		return c.DeleteByPattern(ctx, patterns[0])
	}

	var (
		wg       sync.WaitGroup
		errOnce  sync.Once
		firstErr error
	)
	for _, pattern := range patterns {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			var cursor uint64
			for {
				keys, nextCursor, err := c.client.Scan(ctx, cursor, p, 1000).Result()
				if err != nil {
					errOnce.Do(func() { firstErr = fmt.Errorf("cache scan %q: %w", p, err) })
					return
				}
				if len(keys) > 0 {
					if err := c.client.Del(ctx, keys...).Err(); err != nil {
						errOnce.Do(func() { firstErr = fmt.Errorf("cache delete %q: %w", p, err) })
						return
					}
				}
				cursor = nextCursor
				if cursor == 0 {
					return
				}
			}
		}(pattern)
	}
	wg.Wait()
	return firstErr
}

func (c *Cache) SAdd(ctx context.Context, key string, member string, ttl time.Duration) error {
	pipe := c.client.TxPipeline()
	pipe.SAdd(ctx, key, member)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// SAddMany adds multiple members atomically and refreshes the TTL.
func (c *Cache) SAddMany(ctx context.Context, key string, members []string, ttl time.Duration) error {
	if len(members) == 0 {
		return c.client.Expire(ctx, key, ttl).Err()
	}
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	pipe := c.client.TxPipeline()
	pipe.SAdd(ctx, key, args...)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// MGet returns map[key]rawJSONBytes for hits. Missing keys are absent from the
// returned map. Caller unmarshals per-key (types may differ across keys).
// Falls back to per-key Get on the rare go-redis result-shape mismatch — never
// returns partial errors that hide hits.
func (c *Cache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	vals, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string][]byte, len(keys))
	for i, v := range vals {
		if v == nil {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		out[keys[i]] = []byte(s)
	}
	return out, nil
}

// MSet writes many JSON-encoded values under the same TTL in one pipelined
// round trip. Same semantics as N serial Set calls; just amortizes RTT cost.
func (c *Cache) MSet(ctx context.Context, entries map[string]any, ttl time.Duration) error {
	if len(entries) == 0 {
		return nil
	}
	pipe := c.client.TxPipeline()
	for key, value := range entries {
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		pipe.Set(ctx, key, b, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *Cache) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

func (c *Cache) SRem(ctx context.Context, key string, member string) error {
	return c.client.SRem(ctx, key, member).Err()
}

// SIsMember reports whether a single member exists in the set.
func (c *Cache) SIsMember(ctx context.Context, key, member string) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// SMIsMember reports membership for many members in a single call. Returns booleans in the same order.
func (c *Cache) SMIsMember(ctx context.Context, key string, members []string) ([]bool, error) {
	if len(members) == 0 {
		return nil, nil
	}
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.client.SMIsMember(ctx, key, args...).Result()
}

// Exists reports whether the given key exists.
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (c *Cache) Miss(err error) bool {
	return err == redis.Nil
}

// Incr atomically increments a counter with no expiry. Returns post-increment
// count. Use when counter must persist until explicitly deleted.
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// GetCounter reads an integer counter set by Incr/IncrWithWindow. Returns
// (0, nil) when the key is absent so callers can treat a missing version key
// as version 0 without special-casing redis.Nil. Other errors propagate.
func (c *Cache) GetCounter(ctx context.Context, key string) (int64, error) {
	raw, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return parseInt64(raw)
}

// ZSetMember pairs a sorted-set member with its score for ZAddReplace /
// ZRevRangeWithScores. Member is the application-level identifier (typically
// a stringified ID); Score is the float64 priority Redis sorts by.
type ZSetMember struct {
	Member string
	Score  float64
}

// ZAddReplace atomically replaces a sorted set's contents with the given
// members under one key and applies the TTL. Existing entries are dropped so
// callers can treat the call as a snapshot write. Single MULTI/EXEC pipeline.
func (c *Cache) ZAddReplace(ctx context.Context, key string, members []ZSetMember, ttl time.Duration) error {
	pipe := c.client.TxPipeline()
	pipe.Del(ctx, key)
	if len(members) > 0 {
		zMembers := make([]redis.Z, len(members))
		for i, m := range members {
			zMembers[i] = redis.Z{Score: m.Score, Member: m.Member}
		}
		pipe.ZAdd(ctx, key, zMembers...)
	}
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// ZRevRangeWithScores returns members of a sorted set in score-descending
// order from rank `start` to `stop` inclusive (Redis ZREVRANGE semantics).
// Empty key returns (nil, nil), not an error, so callers can treat absence
// the same as an empty set.
func (c *Cache) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZSetMember, error) {
	z, err := c.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]ZSetMember, len(z))
	for i, x := range z {
		m, _ := x.Member.(string)
		out[i] = ZSetMember{Member: m, Score: x.Score}
	}
	return out, nil
}

// IncrWithWindow atomically increments a counter and ensures it expires after
// the given window. Returns the post-increment count and remaining TTL on the
// key. Use for sliding/rolling rate limits — first hit creates the key with
// the window TTL; subsequent hits only INCR. Single pipelined round trip.
//
// Race: between INCR and the conditional EXPIRE another caller may also see
// "no TTL" and issue EXPIRE again — both set the same window so the last one
// wins. Safe for rate limiting.
func (c *Cache) IncrWithWindow(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	ttl := pipe.TTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, 0, err
	}
	count := incr.Val()
	remaining := ttl.Val()
	if remaining < 0 {
		if err := c.client.Expire(ctx, key, window).Err(); err != nil {
			return count, window, err
		}
		remaining = window
	}
	return count, remaining, nil
}

// CounterTTL reads a counter set by IncrWithWindow without mutating it.
// Returns (0, 0, nil) when the key is absent. Use for "am I over the limit?"
// checks that must not consume a slot.
func (c *Cache) CounterTTL(ctx context.Context, key string) (int64, time.Duration, error) {
	pipe := c.client.Pipeline()
	get := pipe.Get(ctx, key)
	ttl := pipe.TTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return 0, 0, err
	}
	raw, err := get.Result()
	if err == redis.Nil {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	count, err := parseInt64(raw)
	if err != nil {
		return 0, 0, err
	}
	return count, ttl.Val(), nil
}

func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// TagTTL controls how long a reverse-index tag set survives. Must be > the
// TTL of any cache key registered under it; otherwise InvalidateTag risks
// missing keys whose tag set expired first. FacultyTTL == 24h so this is
// bumped to 48h to keep a margin of safety.
const TagTTL = 48 * time.Hour

// RegisterCacheKey adds a cache key to a tag set. Tag sets are used by
// InvalidateTag for O(set size) invalidation instead of O(keyspace) SCAN.
//
// Convention: tag string starts with "tag:" — see tag_keys.go in the
// infrastructure/repository package for the canonical builders.
//
// Side effects: SADD member + refresh TTL on the tag set. Safe to call
// repeatedly for the same (tag, key) pair.
func (c *Cache) RegisterCacheKey(ctx context.Context, tag, cacheKey string) error {
	return c.SAdd(ctx, tag, cacheKey, TagTTL)
}

// RegisterCacheKeyMany registers a cache key under multiple tags in one shot.
// Each tag receives the same key as a member; useful when one cached value
// must be invalidatable by several entity tags (e.g. user feed keyed by
// (user, requester) — register under both).
func (c *Cache) RegisterCacheKeyMany(ctx context.Context, tags []string, cacheKey string) error {
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		if err := c.SAdd(ctx, tag, cacheKey, TagTTL); err != nil {
			return err
		}
	}
	return nil
}

// InvalidateTag drops every cache key registered under the tag, then drops
// the tag set itself. O(members), not O(keyspace) — the win over SCAN.
//
// Pipelined DEL: all members deleted in one round-trip. Returns first error;
// keys not deleted on error remain reachable via SMEMBERS, so re-invalidation
// is idempotent.
func (c *Cache) InvalidateTag(ctx context.Context, tag string) error {
	members, err := c.client.SMembers(ctx, tag).Result()
	if err != nil {
		return fmt.Errorf("cache smembers %q: %w", tag, err)
	}
	if len(members) > 0 {
		if err := c.client.Del(ctx, members...).Err(); err != nil {
			return fmt.Errorf("cache del tag %q: %w", tag, err)
		}
	}
	if err := c.client.Del(ctx, tag).Err(); err != nil {
		return fmt.Errorf("cache del tag set %q: %w", tag, err)
	}
	return nil
}

// InvalidateTags fans InvalidateTag across many tags in parallel. Returns
// first error encountered; other goroutines still complete.
func (c *Cache) InvalidateTags(ctx context.Context, tags ...string) error {
	if len(tags) == 0 {
		return nil
	}
	if len(tags) == 1 {
		return c.InvalidateTag(ctx, tags[0])
	}
	var (
		wg       sync.WaitGroup
		errOnce  sync.Once
		firstErr error
	)
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			if err := c.InvalidateTag(ctx, t); err != nil {
				errOnce.Do(func() { firstErr = err })
			}
		}(tag)
	}
	wg.Wait()
	return firstErr
}
