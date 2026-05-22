package infrarepoShared

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sunibgram/infrastructure/cache"
)

// Viewer-flag Redis sets — per-(user, flag) membership for thread interactions.
// One Redis SET per (user, flag) lets the feed builder do a single SMISMEMBER per flag
// instead of 4 SQL batch queries on every feed request.
//
// Keys:
//   user:{uid}:liked
//   user:{uid}:bookmarked
//   user:{uid}:reposted     (members are original_thread_id for reposts)
//   user:{uid}:pinned
//   user:{uid}:{flag}:hydrated   marker that the set is authoritative (TTL 24h)
//
// Hydration is lazy: on first miss, the set is filled from DB via the supplied loader.

const (
	viewerFlagSetTTL   = 24 * time.Hour
	viewerFlagHydrated = 24 * time.Hour
)

func viewerFlagSetKey(userID uint, flag string) string {
	return fmt.Sprintf("user:%d:%s", userID, flag)
}

func viewerFlagHydratedKey(userID uint, flag string) string {
	return fmt.Sprintf("user:%d:%s:hydrated", userID, flag)
}

// ViewerFlagAdd marks a thread as flagged for a user (e.g. on Like).
// No-op if cache is nil. Safe to call before hydration — SADD creates the set.
func ViewerFlagAdd(c *cache.Cache, userID uint, flag string, threadID uint) {
	if c == nil || userID == 0 || threadID == 0 {
		return
	}
	ctx := context.Background()
	_ = c.SAdd(ctx, viewerFlagSetKey(userID, flag), strconv.FormatUint(uint64(threadID), 10), viewerFlagSetTTL)
}

// ViewerFlagRemove clears a thread from a user's flag set (e.g. on Unlike).
func ViewerFlagRemove(c *cache.Cache, userID uint, flag string, threadID uint) {
	if c == nil || userID == 0 || threadID == 0 {
		return
	}
	ctx := context.Background()
	_ = c.SRem(ctx, viewerFlagSetKey(userID, flag), strconv.FormatUint(uint64(threadID), 10))
}

// ViewerFlagHydrate fills the Redis set from the supplied loader (DB) on first use.
// Caller passes a loader returning the full list of flagged thread_ids for the user.
func ViewerFlagHydrate(c *cache.Cache, userID uint, flag string, loader func() ([]uint, error)) error {
	if c == nil || userID == 0 {
		return nil
	}
	ctx := context.Background()
	hydratedKey := viewerFlagHydratedKey(userID, flag)
	if hit, _ := c.Exists(ctx, hydratedKey); hit {
		return nil
	}
	ids, err := loader()
	if err != nil {
		return err
	}
	setKey := viewerFlagSetKey(userID, flag)
	_ = c.Delete(ctx, setKey)
	members := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		members = append(members, strconv.FormatUint(uint64(id), 10))
	}
	if err := c.SAddMany(ctx, setKey, members, viewerFlagSetTTL); err != nil {
		return err
	}
	return c.Set(ctx, hydratedKey, true, viewerFlagHydrated)
}

// ViewerFlagMembership returns a map[threadID]bool reporting membership for a user's flag set.
// Returns ok=false if the cache is nil or the set hasn't been hydrated — caller must fall back to DB.
func ViewerFlagMembership(c *cache.Cache, userID uint, flag string, threadIDs []uint) (map[uint]bool, bool) {
	if c == nil || userID == 0 || len(threadIDs) == 0 {
		return nil, false
	}
	ctx := context.Background()
	if hit, _ := c.Exists(ctx, viewerFlagHydratedKey(userID, flag)); !hit {
		return nil, false
	}
	members := make([]string, len(threadIDs))
	for i, id := range threadIDs {
		members[i] = strconv.FormatUint(uint64(id), 10)
	}
	results, err := c.SMIsMember(ctx, viewerFlagSetKey(userID, flag), members)
	if err != nil {
		return nil, false
	}
	out := make(map[uint]bool, len(threadIDs))
	for i, id := range threadIDs {
		if i < len(results) && results[i] {
			out[id] = true
		}
	}
	return out, true
}

// ViewerFlagPurge drops all viewer flag sets for a user (e.g. on account deletion).
func ViewerFlagPurge(c *cache.Cache, userID uint) {
	if c == nil || userID == 0 {
		return
	}
	ctx := context.Background()
	flags := []string{"liked", "bookmarked", "reposted", "pinned"}
	keys := make([]string, 0, 2*len(flags))
	for _, f := range flags {
		keys = append(keys, viewerFlagSetKey(userID, f), viewerFlagHydratedKey(userID, f))
	}
	_ = c.Delete(ctx, keys...)
}
