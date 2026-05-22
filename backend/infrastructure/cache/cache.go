package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

func (c *Cache) SAdd(ctx context.Context, key string, member string, ttl time.Duration) error {
	if err := c.client.SAdd(ctx, key, member).Err(); err != nil {
		return err
	}
	return c.client.Expire(ctx, key, ttl).Err()
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
	if err := c.client.SAdd(ctx, key, args...).Err(); err != nil {
		return err
	}
	return c.client.Expire(ctx, key, ttl).Err()
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
