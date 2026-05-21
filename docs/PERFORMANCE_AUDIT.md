# Performance Audit — Thread GET API & Caching Strategy

**Date:** 2026-05-21
**Scope:** Backend (`backend/`) — all endpoints returning thread data, plus Redis cache layer.
**Status:** Findings + proposed solutions + concrete invalidation plan.

---

## 1. Executive Summary

Current GET-thread endpoints work but pay 12–15 SQL queries on cache-miss and 6–8 on cache-hit because viewer-specific data (likes/bookmarks/reposts/pins/polls/`is_premium`/faculty) is fetched per request from Postgres even when the underlying thread list is already cached in Redis.

The biggest wins (in order):

1. **De-fragment feed cache** — split thread list (shared across users) from viewer flags (per-user).
2. **Cache viewer flags** (`liked`/`bookmarked`/`reposted`/`pinned`) per user with `SISMEMBER`-style sets.
3. **Cache polls + per-user vote** (currently zero cache).
4. **Cache `is_premium` + faculty** per user (currently always re-fetched in `FeedBuilder`).
5. **Cursor pagination** instead of `OFFSET` for deep pages (long-term scale).

Items 1–4 are non-breaking, can be implemented incrementally, and each independently cuts request latency 30–60%.

---

## 2. Current State (Baseline)

### 2.1 Per-request cost on `GET /api/threads/` (page=1, limit=20, viewer logged in)

| Stage | Queries | Cached? | Note |
|---|---|---|---|
| Feed list (`threads:roots:viewer:{uid}`) | 1 count + 1 select | ✅ 60s, **viewer-keyed** | 1 entry per user — high fragmentation |
| Thread Preloads (User, Community, Faculty, Events, OriginalThread.\*) | ~6 implicit | implicit in feed cache | re-fired on cache miss |
| Thread media batch | 1 IN | ❌ | indexed, fast |
| Avatar URLs | 1 batch + Redis lookup | ✅ 1h per user | only `is_premium` + faculty miss |
| `is_premium` fresh fetch | 1 batch | ❌ | always re-fetched (LoadAvatarsForUserIDs misses path) |
| Faculty fresh fetch | 1 batch | ❌ | always re-fetched |
| Liked flags | 1 batch | ❌ | `interactionRepo.areThreadsInteracted` |
| Bookmarked flags | 1 batch | ❌ | same |
| Reposted flags | 1 batch | ❌ | same |
| Pinned flags | 1 batch | ❌ | `pinService.ArePinnedFor` |
| **Poll body** (new feature) | 1 polls + 1 options | ❌ | always hits DB |
| **Poll user votes** (new feature) | 1 batch | ❌ | always hits DB |

**Cold cache:** ~14 SQL queries.
**Hot cache (feed hit):** ~10 SQL queries — viewer-specific data always re-fetched.

### 2.2 Cache key fragmentation

| Key | Granularity | Memory cost |
|---|---|---|
| `thread:roots:page:1:limit:20:viewer:42` | **per user** | 1000 active users → 1000 copies of identical thread list |
| `thread:userfeed:{userID}:{tab}:page:..:req:{requesterID}` | **per (target user, viewer)** | quadratic blow-up |
| `thread:usermedia:{userID}:..:req:{requesterID}` | same | same |
| `thread:following:{userID}:page:..` | per user | OK (genuinely viewer-specific) |
| `feed:foryou:{userID}:*` | per user | OK |

Two of these (`thread:roots`, user-feed/media) cache the **same thread rows** for every viewer despite the only viewer-specific difference being privacy filtering. A user-list of public threads is identical across 1000 viewers — should be cached once.

### 2.3 Existing invalidation infrastructure (good)

`backend/infrastructure/repository/thread_cache_invalidation.go` already implements:

- `invalidateFeedCachesByThreadID(threadID)` — via `SAdd` set tracking
- `invalidateForYouFeedCacheForUser(userID)`
- `invalidateForYouFeedCachesForFollowers(userID)`
- `invalidateForYouFeedCachesForCommunity(communityID)`
- `invalidateUserFeedCache`, `invalidateUserMediaCache`, `invalidateFollowingFeedCache`, `invalidateTrendingFeed`
- `interactionRepo.invalidateByThreadID` / `invalidateUser`

**Pattern in use:** on `SET`, `SAdd` the cache key into `thread:feed_keys:{threadID}` so any mutation on that thread can `SMembers` → `Delete` all dependent feed caches. This is correct and reusable for the new caches below.

---

## 3. Improvements — Detailed

### 3.1 De-fragment feed cache (`thread:roots`, user feed, community feed)

**Problem.** Cache keys include `viewer:{id}`. Identical public thread lists are duplicated per user.

**Solution.** Two-tier cache:

```
Layer 1 — thread list (shared):
  feed:roots:public:page:{p}:limit:{l}        → []ThreadID + Total  (TTL 60s)
  feed:community:{cid}:{sort}:page:{p}:limit:{l}
  feed:user:{uid}:{tab}:page:{p}:limit:{l}    (excludes privacy filter)

Layer 2 — thread payload (shared, already exists):
  thread:id:{threadID}                         → model.Thread  (TTL 5min)

Layer 3 — viewer-specific overlay (computed per request, see §3.2):
  user:{vid}:liked, user:{vid}:bookmarked, ...
```

Flow on `GET /threads/`:
1. Read `feed:roots:public:page:1:limit:20` → list of `[]ThreadID`. (1 Redis get)
2. `MGET thread:id:{tid1}` ... → bulk fetch payloads from Redis. (1 Redis mget)
3. Apply privacy filter in Go (read `user_settings` from Redis cache, see §3.4) — skip private accounts the viewer doesn't follow.
4. Overlay viewer flags (§3.2).

**Memory savings.** ~99% for `roots` cache. 1 entry instead of N-users entries.

**Invalidation rules.**

| Trigger | Action |
|---|---|
| Thread created (POST) | `cache.Delete("feed:roots:public:*")` ; `invalidateForYouFeedCachesForFollowers(authorID)` if not community thread |
| Thread updated/deleted | `invalidateFeedCachesByThreadID(threadID)` (already via `SAdd` tracking) + `cache.Delete("thread:id:{tid}")` |
| User changes privacy (private↔public) | `cache.Delete("user_settings:{uid}")` ; **no feed invalidation needed** — filter is applied at read-time using cached settings |
| Follow/unfollow | `cache.Delete("follow_set:{follower}")` ; **no feed invalidation** — filter uses cached follow set |
| Community visibility change | `cache.DeleteByPattern("feed:community:{cid}:*")` ; `invalidateForYouFeedCachesForCommunity(cid)` |

**Risk.** Privacy filter must be cheap at read time. Mitigation: cache `is_private` + follow-set membership in Redis (§3.4).

---

### 3.2 Viewer flags Redis cache

**Problem.** 4 batch queries (like/bookmark/repost/pin) fire on every request even after feed cache hit. For a feed of 20 threads × 4 flags = 80 row lookups per request.

**Solution.** Store viewer flags as Redis sets per user:

```
user:{uid}:liked         → SET of thread_ids the user liked
user:{uid}:bookmarked    → SET
user:{uid}:reposted      → SET (of original_thread_id for reposts)
user:{uid}:pinned        → SET (only the user's own thread_ids)
```

**Read:** for a feed of 20 threads, 1 `SMISMEMBER user:{uid}:liked t1 t2 ... t20` per flag = 4 Redis calls total, sub-ms. Replaces 4 SQL batch queries.

**Hydration (lazy):** on first miss, populate full set from DB. Use `EXISTS` Redis key `user:{uid}:liked:hydrated` to know if set is authoritative or needs to be re-loaded. TTL on hydrated key = 24h to force occasional re-sync.

**Invalidation rules.**

| Trigger | Action |
|---|---|
| `Like(uid, tid)` | `SADD user:{uid}:liked tid` |
| `Unlike(uid, tid)` | `SREM user:{uid}:liked tid` |
| `Bookmark(uid, tid)` | `SADD user:{uid}:bookmarked tid` |
| `Unbookmark(uid, tid)` | `SREM user:{uid}:bookmarked tid` |
| `Repost(uid, originalTid)` | `SADD user:{uid}:reposted originalTid` |
| `Unrepost(uid, originalTid)` | `SREM user:{uid}:reposted originalTid` |
| `Pin(uid, tid)` | `SADD user:{uid}:pinned tid` |
| `Unpin(uid, tid)` | `SREM user:{uid}:pinned tid` |
| Thread deleted | `SREM` from any user set is unnecessary — `SISMEMBER` returns false for deleted threads anyway; orphan set entries are GC'd at next full hydrate |
| User deletes account | `DEL user:{uid}:liked user:{uid}:bookmarked user:{uid}:reposted user:{uid}:pinned user:{uid}:liked:hydrated ...` |

**Where to invalidate (concrete call sites):**

| File | Function | Add |
|---|---|---|
| `infrastructure/repository/like_repository.go` | `Like`/`Unlike` | `cache.SAdd/SRem("user:{uid}:liked", tid)` after DB op |
| `infrastructure/repository/bookmark_repository.go` | `Bookmark`/`Unbookmark` | same pattern |
| `infrastructure/repository/repost_repository.go` | repost create/delete | same pattern |
| `infrastructure/repository/thread_pin_repository.go` | `Pin`/`Unpin` | same pattern |
| `webapi/utils/feed_builder.go` | `LoadViewerFlagsForThreads` | check Redis first; fall back to DB + populate set on miss |

**Hydration helper sketch:**
```go
func (fb *FeedBuilder) loadLikedSet(ctx context.Context, uid uint, threadIDs []uint) map[uint]bool {
    hydratedKey := fmt.Sprintf("user:%d:liked:hydrated", uid)
    if exists, _ := fb.cache.Exists(ctx, hydratedKey); !exists {
        ids, _ := fb.likeSvc.AllLikedThreadIDs(uid) // DB fetch — once per 24h
        fb.cache.SAdd(ctx, fmt.Sprintf("user:%d:liked", uid), toAny(ids)...)
        fb.cache.Set(ctx, hydratedKey, true, 24*time.Hour)
    }
    res, _ := fb.cache.SMIsMember(ctx, fmt.Sprintf("user:%d:liked", uid), threadIDs)
    return res
}
```

**Cost.** ~50KB per active user (avg 1000 liked threads × 4 flags). 10k active users → 500MB Redis. Acceptable. Or use bit-set encoding for 10× reduction.

---

### 3.3 Poll cache (mandatory now — feature has zero caching)

**Problem.** Every thread render w/ poll triggers 2 SQL queries (`thread_polls` + `thread_poll_options`). On a feed of 20, if all threads have polls → 2 batched IN-queries. Per page. No cache.

**Solution.**

```
poll:thread:{threadID}             → full poll w/ options JSON  (TTL 30s)
poll:thread:batch:{hash}           → optional: pre-baked batch, skip
poll:{pollID}:user:{userID}        → optionID the user voted (TTL 5min)
```

Short TTL on `poll:thread:{tid}` because vote counts move. Per-user vote TTL longer because invalidated on mutation.

**Implementation file.** `backend/infrastructure/repository/thread_poll_repository.go`. Add `Get`/`Set` to `FindByThreadID` and `FindByThreadIDs`, and `Get`/`Set` to `FindUserVotes`.

**Invalidation rules.**

| Trigger | Action |
|---|---|
| `CastVote(pollID, optionID, userID)` (incl. switch) | `cache.Delete("poll:thread:{threadID}")` ; `cache.Set("poll:{pollID}:user:{userID}", optionID, 5min)` |
| `Create(poll)` | nothing to invalidate (new threadID) ; but invalidate thread payload `thread:id:{tid}` so feed renders include the poll |
| Thread deleted (cascade) | `cache.Delete("poll:thread:{tid}")` ; `cache.DeleteByPattern("poll:{pollID}:user:*")` |
| Poll expires (passive) | TTL handles it ; no explicit invalidation needed |

**Where:** wire invalidation into existing methods in `thread_poll_repository.go`:

- `CastVote` end of transaction → after commit, `cache.Delete(pollThreadKey(threadID))` and `cache.Set(userVoteKey(pollID, userID), optionID, 5*time.Minute)`. Needs poll → threadID lookup, already available since we fetched poll at start of vote service.
- `DeleteByThreadID` → already invalidates by pattern, good.
- `Create` → invalidate `thread:id:{tid}` because thread payload should now include poll on next read.

**Code change required:** `Vote` in `thread_poll_service.go` must pass `threadID` (not just pollID) to repo, or repo must derive it. Cleanest: pass both.

---

### 3.4 `is_premium` + faculty + privacy + follow-set caches

**Problem.** `LoadAvatarsForUserIDs` fires `GetUsers` + `GetUserFaculties` on every cache miss because thread cache may be stale w.r.t. premium status (user upgraded premium after thread was cached). Same problem for faculty changes.

**Solution.** Cache these per user in Redis with explicit invalidation:

```
user:{uid}:profile_meta     → { is_premium, faculty_id, faculty_name }  (TTL 24h)
user:{uid}:is_private       → bool  (TTL 24h)
follow_set:{follower_id}    → SET of following_ids  (TTL 24h, lazy-hydrate)
```

**Invalidation rules.**

| Trigger | Action |
|---|---|
| Premium subscription activated/expired | `cache.Delete("user:{uid}:profile_meta")` |
| Faculty changed | `cache.Delete("user:{uid}:profile_meta")` |
| Avatar changed | `cache.Delete("user:{uid}:avatar")` (already exists) ; `cache.Delete("user:{uid}:profile_meta")` |
| Privacy toggled | `cache.Delete("user:{uid}:is_private")` |
| Follow/unfollow | `SADD/SREM follow_set:{follower}` ; `cache.Set("follow_set:{follower}:hydrated", true, 24h)` |
| User deletes account | drop all `user:{uid}:*` keys |

**Where:**

| File | Function | Action |
|---|---|---|
| `infrastructure/repository/premium_subscription_repository.go` | Activate / Expire | delete `user:{uid}:profile_meta` |
| `infrastructure/repository/user_faculty_repository.go` | Set | delete `user:{uid}:profile_meta` |
| `infrastructure/repository/user_settings_repository.go` | SetPrivacy | delete `user:{uid}:is_private` |
| `infrastructure/repository/user_profile_repository.go` | Update (avatar) | delete `user:{uid}:profile_meta` |
| `infrastructure/repository/follow_repository.go` | Follow / Unfollow | `SADD`/`SREM follow_set:{follower}` |

---

### 3.5 Cursor pagination (long-term)

**Problem.** `OFFSET 1000 LIMIT 20` on `threads` table scans 1020 rows then discards 1000. At 100k+ threads, deep-page latency degrades to seconds.

**Solution.** Cursor (keyset) pagination:

```sql
WHERE (created_at, thread_id) < (?, ?)
ORDER BY created_at DESC, thread_id DESC
LIMIT 20
```

**API change.** Replace `?page=2` with `?cursor=<base64(created_at:thread_id)>`. Existing index `idx_threads_created_at` already supports this directly.

**Migration.** Keep `page`-based endpoints alive; add new `?cursor=` variant. Frontend can migrate gradually. Eventually deprecate `page`.

**Caching impact.** Cache key changes from `page:1:limit:20` to `cursor:abc:limit:20`. Same TTL applies.

---

### 3.6 Replace some `Preload` with `JOIN` (later, measure first)

**Problem.** `preloadThreadDetails` issues 8 implicit `SELECT ... WHERE id IN (...)` queries (User, Community, Faculty, Events, OriginalThread.\*). For 20 threads that's 8 small queries with overhead.

**Solution.** Switch hot fields (`User.name`, `User.user_name`, `User.is_premium`) into a single `LEFT JOIN users` on the thread query, projecting columns. Reduces 8 round-trips to 1.

**Trade-off.** Complex queries are harder to maintain and reason about. Profile first with `EXPLAIN ANALYZE` and verify the bottleneck is round-trips, not result size, before refactoring.

**Recommendation.** Only do this if `pg_stat_statements` shows preload queries in the top 10 by total time. Otherwise leave alone.

---

### 3.7 Cursor / batch poll loader

**Already done ✅** — `FindByThreadIDs` + `FindUserVotes` already batched. No further change needed at SQL layer once §3.3 caches added.

---

## 4. Concrete Invalidation Matrix (full project scan)

This table consolidates every mutation across the backend and the cache keys it must invalidate. Source files referenced from `backend/infrastructure/repository/`.

| Mutation | Source file | Cache keys to invalidate |
|---|---|---|
| Thread created | `thread_repository_write.go::Save` | `feed:roots:public:*` ; `thread:userfeed:{authorID}:*` ; `thread:usermedia:{authorID}:*` ; `feed:foryou:{follower}:*` for each follower ; `feed:trending:global` ; if community → `feed:community:{cid}:*` + foryou for community members |
| Thread updated | `thread_repository_write.go::Update` | `thread:id:{tid}` ; via `SMembers thread:feed_keys:{tid}` → delete all tracked keys |
| Thread deleted | `thread_repository_write.go::Delete` | same as updated + `thread:feed_keys:{tid}` (the tracking key itself) + `poll:thread:{tid}` + `poll:{pollID}:user:*` (via `DeleteByPattern`) |
| Like / Unlike | `like_repository.go` | `SADD/SREM user:{uid}:liked tid` ; patch `thread:id:{tid}.like_count` in Redis ; via `interactionRepo.invalidateByThreadID(tid)` (existing) |
| Bookmark / Unbookmark | `bookmark_repository.go` | `SADD/SREM user:{uid}:bookmarked tid` ; patch counter ; `invalidateByThreadID` |
| Repost / Unrepost | `repost_repository.go` | `SADD/SREM user:{uid}:reposted originalTid` ; patch counter on original ; `invalidateByThreadID(originalTid)` |
| Pin / Unpin | `thread_pin_repository.go` | `SADD/SREM user:{uid}:pinned tid` ; `thread:userfeed:{uid}:*` (pins re-order user feed) |
| Reply created | `thread_repository_write.go::Save` (Type=REPLY) | `thread:chain:{rootID}:*` ; `thread:id:{parentID}` (comment_count++) |
| Repost created | `thread_repository_write.go::Save` (Type=REPOST) | feed caches for follower for-you + `thread:userfeed:{reposterID}` |
| Follow / Unfollow | `follow_repository.go` | `SADD/SREM follow_set:{follower} following` ; `cache.Delete("thread:following:{follower}:*")` ; `feed:foryou:{follower}:*` |
| Community join / leave | `community_member_repository.go` | `feed:foryou:{userID}:*` ; community member cache |
| Community thread created | `thread_repository_write.go::Save` (with CommunityID) | `feed:community:{cid}:*` ; `feed:foryou:{memberID}:*` for each member |
| Premium activated / expired | `premium_subscription_repository.go` | `user:{uid}:profile_meta` (forces feed renders to re-fetch `is_premium`) |
| Faculty set / changed | `user_faculty_repository.go` | `user:{uid}:profile_meta` ; `user:{uid}:avatar` (no — avatar unaffected) |
| Privacy toggled | `user_settings_repository.go::SetPrivacy` | `user:{uid}:is_private` ; **do not** invalidate feed caches — privacy is applied at read-time using cached flag |
| Avatar changed | `user_profile_repository.go::Update` | `user:{uid}:avatar` (existing) ; `user:{uid}:profile_meta` |
| Poll created | `thread_poll_repository.go::Create` | `thread:id:{tid}` (so feed includes poll on next read) |
| Vote cast / switched | `thread_poll_repository.go::CastVote` | `poll:thread:{tid}` ; `poll:{pollID}:user:{uid}` (re-set, not delete) |
| Poll deleted (with thread) | `thread_poll_repository.go::DeleteByThreadID` | `poll:thread:{tid}` ; `poll:{pollID}:user:*` (pattern) |
| Hashtag-counter update | `hashtag_repository.go` | `feed:trending:global` (existing) |
| Event-counter update | `event_repository.go` | `event:trending:global` if exists |
| User account deleted | `user_repository.go::Delete` | `user:{uid}:*` (broad pattern) ; `follow_set:*` containing this user — relax: TTL 24h handles cleanup ; thread payloads stay (orphan but harmless until thread itself is deleted) |

---

## 5. Implementation Roadmap

### Phase 1 — High ROI, low risk (1–2 days)

1. **Poll Redis cache** (§3.3)
   - `poll:thread:{tid}` 30s TTL on `FindByThreadID` / `FindByThreadIDs`
   - `poll:{pollID}:user:{uid}` 5min TTL on `FindUserVotes`
   - Invalidate on `CastVote`, `Create`, `DeleteByThreadID`
   - **Files touched:** `thread_poll_repository.go`, `thread_poll_service.go` (pass threadID to repo)

2. **Counter patches via Redis** (already partially done for likes/bookmarks via `patchCount`)
   - Extend to `bookmark_count`, `repost_count`, `comment_count` if not already.

### Phase 2 — Big wins, moderate effort (3–5 days)

3. **Viewer flag sets** (§3.2)
   - Lazy-hydrate per (user, flag) pair with `:hydrated` marker
   - `SADD/SREM` in like/bookmark/repost/pin repos
   - `SMISMEMBER` in `FeedBuilder.LoadViewerFlagsForThreads`
   - Add `AllLikedThreadIDs`, `AllBookmarkedThreadIDs`, `AllRepostedOriginalIDs`, `AllPinnedThreadIDs` to respective services for hydration

4. **De-fragment feed cache** (§3.1)
   - Split `thread:roots` → `feed:roots:public` (shared) + viewer overlay
   - Apply privacy filter in Go using cached `user:{uid}:is_private` + `follow_set:{follower}`
   - Drop `viewer:` suffix from cache keys
   - Same pattern for `thread:userfeed`, `thread:community`

### Phase 3 — Profile + measure before doing (1 week+)

5. **Profile-meta cache** (§3.4)
6. **Cursor pagination** (§3.5) — only if `OFFSET` shows up in slow-query log
7. **Preload → JOIN** (§3.6) — only if preload queries top the slow-query list

---

## 6. Monitoring + Verification

Before each phase ships:

1. **Capture baseline** — `pg_stat_statements` top 20 by total time, before changes.
2. **Add Redis hit-rate counter** — `redis.info().keyspace_hits / (hits+misses)`. Expect >95% for thread payloads, >99% for profile meta.
3. **End-to-end latency** — p50, p95, p99 on `GET /api/threads/`, `GET /api/threads/following`, `GET /api/threads/:id/chain` before/after.

Target after Phase 1+2:
- p50 < 25ms (cache hit, viewer logged in)
- p95 < 80ms
- p99 < 200ms (cold cache including poll batch)
- SQL queries per request: 2–4 (down from 10–14)

---

## 7. Risks + Mitigations

| Risk | Mitigation |
|---|---|
| Cache stampede on TTL expiry of `feed:roots:public` | Use "soft TTL" pattern: serve stale on miss, re-fetch async. Or singleflight in repo (one goroutine refreshes, others wait). |
| Redis sets grow unbounded for power users | TTL 24h on hydrated marker forces re-sync; in re-sync, replace set instead of adding (`DEL` then `SADD`). |
| Invalidation miss → stale data shown | Each mutation must hit invalidation. Code-review checklist: any new mutation method MUST add invalidation call. Add a lint rule or test that asserts every write to a cached entity has a matching `cache.Delete`. |
| Cache memory blow-up | Monitor `INFO memory` ; set `maxmemory-policy allkeys-lru` ; consider compressing thread payloads with `msgpack` (~30% smaller than JSON). |
| Privacy filter at read-time leaks data | Unit tests: assert private-account threads never appear in `feed:roots:public` results for non-followers. Add integration tests covering edge cases (just-toggled-to-private, just-unfollowed). |

---

## 8. Out of Scope (intentional)

- **Database read replicas** — already split write/read DB ✅, no further action.
- **CDN for media URLs** — MinIO public bucket policy already covers static asset caching at the edge.
- **GraphQL / response shape changes** — keep REST contract stable; this audit is internal-only.
- **Frontend changes** — none required; cache layer is transparent.

---

## 9. Quick Reference — Files To Touch

| Phase | File | Change |
|---|---|---|
| 1 | `backend/infrastructure/repository/thread_poll_repository.go` | Add Redis Get/Set in find methods + invalidation in CastVote/Delete |
| 1 | `backend/application/usecase/thread_poll_service.go` | Pass threadID through to repo CastVote so invalidation key is computable |
| 2 | `backend/webapi/utils/feed_builder.go` | Replace `LoadViewerFlagsForThreads` SQL path with Redis `SMISMEMBER` (fall back to DB on hydration) |
| 2 | `backend/infrastructure/repository/like_repository.go` | `SADD/SREM user:{uid}:liked` on Like/Unlike |
| 2 | `backend/infrastructure/repository/bookmark_repository.go` | same pattern |
| 2 | `backend/infrastructure/repository/repost_repository.go` | same pattern |
| 2 | `backend/infrastructure/repository/thread_pin_repository.go` | same pattern |
| 2 | `backend/domain/service/i_like_service.go` (+ bookmark/repost/pin) | Add `AllXxxThreadIDs(userID)` for hydration |
| 2 | `backend/infrastructure/repository/thread_repository_query.go` | Drop `viewer:` from cache keys ; add privacy filter at read |
| 2 | `backend/infrastructure/repository/thread_cache_invalidation.go` | Update key helpers to remove viewer suffix |
| 3 | `backend/infrastructure/repository/premium_subscription_repository.go` | Invalidate `user:{uid}:profile_meta` |
| 3 | `backend/infrastructure/repository/user_faculty_repository.go` | same |
| 3 | `backend/infrastructure/repository/user_settings_repository.go` | Invalidate `user:{uid}:is_private` |
| 3 | `backend/infrastructure/repository/follow_repository.go` | `SADD/SREM follow_set:{follower}` |

---

## 10. Verification Checklist Before Merging Each Phase

- [ ] All new cache keys are documented in `thread_cache_invalidation.go` (or new `*_cache_invalidation.go` file).
- [ ] Every mutation path that affects a cached entity has a corresponding invalidation call.
- [ ] Integration test covering: mutation → read sees fresh value within 1 request cycle.
- [ ] `go vet ./...` and `go build ./...` pass.
- [ ] Manual smoke test of every affected endpoint (list at end of §2.1 plus poll endpoints).
- [ ] Redis hit rate > 80% after warm-up.
- [ ] No regression in p95 latency.

---

**End of audit.**
