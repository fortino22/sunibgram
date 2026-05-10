# Thread Query Cost Analysis

> Sorted cheapest → most expensive. Includes root cause and fix for each tier.

---

## Cost Tiers

### Low — Single or cached queries

| API / Method | Route | DB Queries | Notes |
|---|---|---|---|
| `ExistsRepost` | internal | 1 COUNT | No joins |
| `ExistsPin` | internal | 1 COUNT | No joins |
| `DeleteRepost` | internal | SELECT + DELETE | 2 queries max |
| `DeleteByIDs` (media) | internal | SELECT + DELETE | 2 queries max |
| `Pin` / `Unpin` | internal | 1 INSERT / DELETE | No joins |
| `FindByThreadID` | internal | 1 SELECT | No joins |
| `FindByThreadIDs` | internal | 1 SELECT IN | Batch, no joins |
| `SaveAll` (media) | internal | 1 batch INSERT | No joins |
| `FindByCommunityID` | `GET /communities/:id/threads` | COUNT + SELECT | No joins, **Redis cached** |
| `FindByID` | `GET /threads/:id` | 1 SELECT + 4 Preloads | **Redis cached**, `attachFaculties` = 1 extra query |

**Fix needed:** `FindByID` calls `attachFaculties` even on cache hit — wasted query when thread is already cached.

```go
// Current (wasteful on cache hit):
if err := r.cache.Get(ctx, key, &thread); err != nil {
    // DB fetch + attachFaculties
} else {
    r.db.attachFaculties(threads) // ← unnecessary, data came from cache
}

// Fix: embed faculty in cached struct, skip attachFaculties on hit
```

---

### Medium — Paginated with 1 join + Redis cache

| API / Method | Route | DB Queries | Notes |
|---|---|---|---|
| `FindRoots` | `GET /threads` | COUNT + SELECT + attachFaculties | 1 LEFT JOIN communities, **Redis cached** |
| `FindByRootID` | `GET /threads/:id/chain` | COUNT + SELECT + attachFaculties | 1 LEFT JOIN communities, **Redis cached** |

**Query pattern:**
```sql
-- Count
SELECT COUNT(*) FROM threads
LEFT JOIN communities ON communities.community_id = threads.community_id
WHERE threads.type IN ('post','repost')
AND (threads.community_id IS NULL OR communities.visibility = 'public')

-- Fetch (same WHERE, adds ORDER BY + OFFSET LIMIT + Preloads)
```

**Fix:** These are already cached — low urgency. Add index on `threads.created_at DESC` if not present.

---

### High — No cache, 2 joins, N+1 via `attachFaculties`

| API / Method | Route | DB Queries | Notes |
|---|---|---|---|
| `FindPostsByUserID` | `GET /users/:id/threads` | COUNT + SELECT + attachFaculties + parent IN | Delegates to `findUserFeed` |
| `FindReplyTabByUserID` | `GET /users/:id/replies` | COUNT + SELECT + attachFaculties + parent IN | Delegates to `findUserFeed` |
| `findUserFeed` (base) | — | **4 queries minimum** | 2 LEFT JOINs, no cache |

**Query breakdown for `findUserFeed`:**
```
1. COUNT query     — LEFT JOIN communities + community_members
2. SELECT threads  — same joins + 4 Preloads (User, OriginalThread.User, Community, OriginalThread.Community)
3. attachFaculties — SELECT user_faculties WHERE user_id IN (...)  ← extra round trip
4. parent threads  — SELECT threads WHERE thread_id IN (parent IDs)
[5. attachFaculties on parents — another SELECT user_faculties]
```

**Fix:** Collapse `attachFaculties` into main SELECT using JOIN instead of post-fetch lookup:
```sql
-- Instead of separate attachFaculties query, add to main query:
LEFT JOIN user_faculties uf ON uf.user_id = threads.user_id
LEFT JOIN faculties f ON f.faculty_id = uf.faculty_id
```

In GORM:
```go
r.db.readDB.
    Preload("User", preloadThreadUser).
    Preload("User.UserFaculty.Faculty").   // ← replaces attachFaculties call
    Preload("OriginalThread.User", preloadThreadUser).
    Preload("OriginalThread.User.UserFaculty.Faculty").
    ...
```

Also add **Redis cache** with short TTL (60s) keyed by `user:{id}:posts:page:{p}:limit:{l}`.

---

### Very High — Multi-phase fetch, no cache, highest join complexity

| API / Method | Route | DB Queries | Notes |
|---|---|---|---|
| `FindMediaByUserID` | `GET /threads/media/:userId` | **5 queries** | 2-phase: ID fetch → full fetch + attachFaculties × 2 |
| `FindFollowingFeed` | `GET /threads/following` | **4–5 queries** | 3 JOINs including `user_relationships` (large table) |

#### `FindMediaByUserID` — 5-query chain

```
1. COUNT DISTINCT threads.thread_id  — JOIN thread_media + community_members
2. SELECT thread IDs (GROUP BY + ORDER BY MAX)
3. SELECT full threads WHERE thread_id IN (ids)
4. attachFaculties(threads)
5. SELECT parent threads WHERE thread_id IN (parent IDs)
[6. attachFaculties(parentThreads)]
```

**Fix:** Collapse phase 2+3 into one query using window function or subquery. Eliminate `attachFaculties` with Preload chain:
```go
// Replace 2-phase ID → fetch with single query:
r.db.readDB.
    Joins("JOIN thread_media ON thread_media.thread_id = threads.thread_id").
    Joins("LEFT JOIN community_members ON ...").
    Preload("User.UserFaculty.Faculty").
    Where(...).
    Group("threads.thread_id").
    Order("MAX(threads.created_at) DESC").
    Offset(offset).Limit(limit).
    Find(&threads)
```

Add Redis cache keyed by `user:{id}:media:page:{p}:limit:{l}`, TTL 60s.

#### `FindFollowingFeed` — highest real-world cost

```
1. COUNT  — JOIN user_relationships + communities + community_members
2. SELECT — same 3 JOINs + 4 Preloads
3. attachFaculties(threads)
4. SELECT parent threads IN (ids)
[5. attachFaculties(parentThreads)]
```

`user_relationships` is the largest table and has no size bound. As followers grow, both COUNT and SELECT scans grow linearly.

**Fix (ordered by impact):**
1. Add composite index: `CREATE INDEX idx_user_rel_follower_following ON user_relationships(follower_id, following_id)`
2. Replace `attachFaculties` with `Preload("User.UserFaculty.Faculty")`
3. Replace COUNT + SELECT (double pass) with single paginated query + cached total
4. Add Redis cache, TTL 30s, keyed by `user:{id}:following:page:{p}:limit:{l}`

---

## Root Cause: `attachFaculties` Extra Round Trip

Every feed function calls this after main fetch:

```go
// thread_repository.go:48
db.readDB.Preload("Faculty").Where("user_id IN ?", ids).Find(&faculties)
```

This fires **one extra SELECT per call**, after the main query already returned. With parent threads, it fires **twice**. Fix once, applies everywhere:

```go
// In preload chain, replace attachFaculties with:
Preload("User.UserFaculty.Faculty").
Preload("OriginalThread.User.UserFaculty.Faculty")
```

Delete `attachFaculties` function after all callers migrated.

---

## Recommended Index Additions

```sql
-- Following feed (most critical)
CREATE INDEX idx_user_rel_follower ON user_relationships(follower_id);

-- Thread feed ordering
CREATE INDEX idx_threads_created_at ON threads(created_at DESC);
CREATE INDEX idx_threads_user_type ON threads(user_id, type);

-- Media feed
CREATE INDEX idx_thread_media_thread_id ON thread_media(thread_id);
```

---

## Priority Order

| Priority | Target | Expected Gain |
|---|---|---|
| 1 | Add `user_relationships` index | Following feed: O(N) → O(log N) scan |
| 2 | Replace `attachFaculties` with Preload | Remove 1–2 extra round trips per feed call |
| 3 | Add Redis cache to `findUserFeed` | Eliminate repeated DB hits for same page |
| 4 | Collapse `FindMediaByUserID` 2-phase into 1 query | Remove 1 extra SELECT |
| 5 | Fix `FindByID` cache-hit calling `attachFaculties` | Minor, low traffic |
