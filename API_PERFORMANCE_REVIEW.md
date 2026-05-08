# API Performance Review — Sunibgram Backend

_Generated: 2026-05-08_

---

## Highest Cost Endpoints (ranked by query load)

| Rank | Endpoint | Queries/Request | Caching | Issues |
|------|----------|----------------|---------|--------|
| 1 | `GET /threads/following` | 4–6 | Partial | 3 JOINs + cache stampede |
| 2 | `POST /chats/messages/read` | 1 + N per msg | None | **N+1 loop query** |
| 3 | `GET /threads/:id/chain` | 3–5 | Partial | Missing pagination index |
| 4 | `GET /likes` / `GET /bookmarks` | 2–3 | Yes | Batch preload missing |
| 5 | `GET /communities/:id/members` | 2 | None | No composite index |
| 6 | `GET /profile/:id` | 2–3 | None | No caching at all |
| 7 | `GET /follows/:user_id/following` | 2 | None | User data not cached |
| 8 | `GET /chats/rooms/:id/messages` | 2 | None | Soft delete overhead |

---

## Critical Issues (fix these first)

### 1. N+1 Query — Chat Read Marking
**File:** `backend/application/usecase/chat_service.go` ~line 108  
**Severity:** HIGH — fires on every mark-read call

```go
// BAD — 1 COUNT query per message
for _, chatID := range inserted {
    if s.chatReadRepo.CountReads(chatID) >= memberCount-1 {
        _ = s.chatRepo.UpdateIsRead(chatID)
    }
}
```

**Fix:** batch the count query
```go
// GOOD — 1 query for all IDs
readCounts, _ := s.chatReadRepo.CountReadsBatch(ctx, inserted)
for chatID, count := range readCounts {
    if count >= memberCount-1 {
        _ = s.chatRepo.UpdateIsRead(chatID)
    }
}
```
Add to repo:
```go
func (r *chatReadRepo) CountReadsBatch(ctx context.Context, chatIDs []uint) (map[uint]int64, error) {
    type result struct {
        ChatID uint
        Count  int64
    }
    var rows []result
    r.db.Table("chat_reads").
        Select("chat_id, COUNT(*) as count").
        Where("chat_id IN ?", chatIDs).
        Group("chat_id").
        Scan(&rows)
    m := make(map[uint]int64, len(rows))
    for _, r := range rows { m[r.ChatID] = r.Count }
    return m, nil
}
```

---

### 2. Cache Stampede — Pattern Invalidation
**File:** `backend/infrastructure/repository/thread_cache_invalidation.go` ~line 29  
**Severity:** HIGH — nukes all paginated caches on any thread write

```go
// BAD — deletes ALL root caches simultaneously
r.cache.DeleteByPattern(ctx, "thread:roots:*")
r.cache.DeleteByPattern(ctx, "thread:chain:*")
```

**Fix options (pick one):**
- Use short TTL (60s) instead of explicit invalidation — stale-but-fast
- Use versioned cache keys: `thread:roots:v{version}:{page}` — bump version on write, old keys expire naturally
- Selective invalidation: only delete specific chain IDs, not wildcard

---

### 3. Missing Composite Indexes — Join Tables
**Severity:** HIGH — full table scans on every existence check

```sql
-- Add these to your migration

-- Interaction tables
CREATE INDEX idx_likes_user_thread ON likes(user_id, thread_id);
CREATE INDEX idx_bookmarks_user_thread ON bookmarks(user_id, thread_id);
CREATE INDEX idx_reposts_user_thread ON reposts(user_id, thread_id);

-- Thread feed pagination
CREATE INDEX idx_thread_user_type ON threads(user_id, type, created_at DESC);
CREATE INDEX idx_thread_feed ON threads(type, community_id, created_at DESC);

-- User relationship (following feed JOIN)
CREATE UNIQUE INDEX idx_user_relationship ON user_relationships(follower_id, following_id);
CREATE INDEX idx_user_relationship_following ON user_relationships(following_id);

-- Community membership checks
CREATE UNIQUE INDEX idx_community_member ON community_members(community_id, user_id);
CREATE INDEX idx_community_creator ON communities(user_id, created_at DESC);

-- Chat membership checks
CREATE UNIQUE INDEX idx_chat_room_member ON chat_room_members(room_id, user_id);
CREATE INDEX idx_chat_room_deleted ON chats(room_id, deleted_at, created_at DESC);

-- Orders
CREATE INDEX idx_order_status ON orders(status, user_id);
CREATE INDEX idx_order_user ON orders(user_id, created_at DESC);
```

---

### 4. Heavy Following Feed Query
**File:** `backend/infrastructure/repository/thread_repository.go` ~line 350  
**Severity:** HIGH — most frequently hit endpoint

```go
// 3 JOINs + multiple WHERE conditions + 4 Preloads
r.readDB.
    Preload("User").Preload("OriginalThread.User").
    Preload("Community").Preload("OriginalThread.Community").
    Joins("JOIN user_relationships ON ...").
    Joins("LEFT JOIN communities ON ...").
    Joins("LEFT JOIN community_members ON ...").
    Where("user_relationships.follower_id = ?", userID).
    Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?", ...).
    Find(&threads)
```

**Fixes:**
1. Add indexes listed in §3 — biggest impact, least code change
2. Limit Preload fields:
```go
Preload("User", func(db *gorm.DB) *gorm.DB {
    return db.Select("user_id", "name", "user_name", "profile_pic_url")
})
```
3. Cache the following-set per user (already partially done) — extend to community membership too

---

## Medium Issues

### 5. No Caching on User Profiles
**File:** `backend/infrastructure/repository/user_repository.go`  
User data is fetched repeatedly (preloaded into every thread). Add Redis cache matching pattern in `thread_repository.go`.

### 6. Cache Invalidation Loop — Interaction Tables
**File:** `backend/infrastructure/repository/interaction_query.go` ~line 46  

When thread is updated, loops over all user IDs who interacted to delete their cache keys. 100 likes = 100 cache deletes.

**Fix:** Use single wildcard key or store interaction caches under thread key, not user key.

### 7. No Privacy Setting Cache
**File:** `backend/application/usecase/thread_service.go` ~line 243  

`settingsRepo.GetPrivacyMapByUserIDs()` called on every feed render. Add Redis TTL cache (5 min) on user privacy settings.

### 8. Community Operations Not Transactional
**File:** `backend/application/usecase/community_service.go` ~line 83  

Community create saves community then member separately — orphaned member possible on partial failure.

```go
// Wrap in transaction
tx := s.db.Begin()
// create community with tx
// add creator as member with tx
tx.Commit()
```

---

## Low Issues

| Issue | File | Impact |
|-------|------|--------|
| SELECT * on User preloads | thread_repository.go | Low — network overhead |
| No index on order.status | order_repository.go | Low — infrequent query |
| Redis/DB race on counter update | interaction_query.go | Low — eventual consistency |
| Thread type column no index | threads table | Low — usually combined with user_id |

---

## Recommended Fix Order

1. **Add all missing indexes** (migration file, zero code change, huge impact)
2. **Fix chat read N+1** (chat_service.go, ~20 lines)
3. **Fix cache stampede** (cache_invalidation.go, TTL approach is simplest)
4. **Limit Preload field selection** (thread_repository.go, repeated pattern)
5. **Cache user profiles** (user_repository.go, copy thread caching pattern)
6. **Wrap community create in transaction** (community_service.go)
7. **Cache privacy settings** (thread_service.go)
