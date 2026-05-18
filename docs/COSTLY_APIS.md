# Costly APIs — Backend Optimization Report

Scan of `backend/` for expensive endpoints. Each item: where, why costly, cheap fix (no logic change).

Priority legend: **P0** = ship first, **P1** = next, **P2** = nice-to-have.

---

## P0 — Fix First

### 1. N+1 Viewer Flags in Every Feed
**Where:** [feed_builder.go:101-123](webapi/utils/feed_builder.go#L101-L123), called from all feed endpoints (GetRoots, GetFollowingFeed, GetMyPosts, GetMyReplies, Explore/GetTop).

**Cost:** `viewerFlags()` runs per thread. Calls `IsLiked`, `IsBookmarked`, `IsReposted`, `IsPinned` → 4 DB roundtrips per thread. 20-thread feed = **80 queries**. Reposts double it.

**Fix:**
- Add batch methods on each service:
  ```go
  AreThreadsLiked(userID uint, threadIDs []uint) map[uint]bool
  AreThreadsBookmarked(userID uint, threadIDs []uint) map[uint]bool
  AreThreadsReposted(userID uint, threadIDs []uint) map[uint]bool
  ArePinnedFor(userID uint, threadIDs []uint) map[uint]bool
  ```
- Each = single `SELECT thread_id FROM likes WHERE user_id=? AND thread_id IN (?)`.
- `feed_builder` collects thread IDs first, calls 4 batch queries, then fills flags from maps.
- Result: 80 queries → **4 queries**.

---

### 2. Unbounded 10K Thread Fetch in `GetChain`
**Where:** [thread_service_query.go:52](application/usecase/thread_service_query.go#L52) — `s.threadRepo.FindByRootID(trueRootID, 1, 10000)`.

**Cost:** Loads up to 10,000 rows into memory, filters and paginates in Go. OOM risk on deep threads.

**Fix:**
- Push pagination into SQL: `FindByRootID(rootID, page, limit)` where `limit` is the user's real page size (default 20).
- If visibility filtering needed, do it in SQL `WHERE` (see #10) instead of post-filter.
- Order in DB (`ORDER BY created_at`) not in Go.

---

### 3. Unbounded 10K + Sync Ollama in Thread Summary
**Where:** [thread_summary_service.go:62, 76](application/usecase/thread_summary_service.go#L62) — `FindByRootID(threadID, 1, 10000)` + `ollamaClient.StreamGenerate(...)` on request path.

**Cost:** Loads 10K rows → trims to top 30 in Go. Then blocks HTTP request for 5-30s on Ollama call. Client sees timeout.

**Fix (two parts, neither change logic):**
1. Query top N in SQL: `ORDER BY like_count DESC, created_at DESC LIMIT 50`. 10K rows → 50.
2. Async Ollama:
   - On request: check Redis `thread:summary:{id}`. Hit → return. Miss → enqueue job to background goroutine/worker, return `202 Accepted` or last cached summary.
   - Worker writes result to Redis with TTL (e.g. 1h). Same summary served to next callers.
   - Add singleflight per thread ID so 100 concurrent requests trigger one Ollama call.

---

### 4. Per-Notification Sender Profile Loads
**Where:** [notification_service.go:98](application/usecase/notification_service.go#L98) — `senderProfile()` per row → 3 queries (user + profile + faculty).

**Cost:** 20 notifications × 3 = **60 queries** per `GET /notifications`.

**Fix:**
- Collect distinct `SenderUserID` set from the page.
- Batch fetch:
  - `userRepo.FindByIDs(ids)`
  - `profileRepo.FindByUserIDs(ids)`
  - `userFacultyRepo.FindByUserIDs(ids)`
- Build `map[userID]senderDTO`, fill each notification from map.
- 60 queries → **3 queries**.

---

## P1 — Async the Blocking I/O

### 5. Sync SMTP Send on Request
**Where:** [email_controller.go:38](webapi/controller/email_controller.go#L38) — `c.emailService.SendMail()` blocks 1-10s.

**Fix:** Push to Redis list / channel. Background worker drains list, calls SMTP. Controller returns `202 {"message":"email queued"}` immediately. No retry logic change — worker retries same way.

---

### 6. Sync MinIO Uploads on Request
**Where:** Thread/chat controllers with media (e.g. [thread_controller.go:59](webapi/controller/thread_controller.go#L59), [chat_controller.go:82](webapi/controller/chat_controller.go#L82)).

**Cost:** Each MinIO `PutObject` blocks request. Multiple files = serial uploads.

**Fix:**
- Parallel uploads in same request: spawn one goroutine per file, `sync.WaitGroup` to join. Same total bytes but wall-clock = max(file_times) not sum.
- Or: write metadata row with `status=uploading`, return immediately, async upload in worker, push WS event `media_ready` to user. Frontend swaps placeholder.

---

### 7. No Cache on Hot Read Endpoints
**Where:**
- `ExploreSearch` — search by query, no Redis.
- `GetTopPosts`, `GetLatestPosts` ([explore endpoints](webapi/router/explore_routes.go)) — no cache.
- Feed visibility helper `GetPrivacyMapByUserIDs`, `GetFollowingSet` called every request.

**Fix:**
- Key pattern: `explore:top:{query}:{page}:{limit}` TTL 60s. Same for latest.
- Search: `explore:search:{hash(query)}:{page}` TTL 30-60s. Privacy filter post-cache (cheap with #1+#10).
- `user:privacy:{id}` cache per user, TTL 5min, invalidate on settings PATCH.
- `user:following:{id}` cache, TTL 5min, invalidate on follow/unfollow.

---

## P2 — Smaller Wins

### 8. Cache Invalidation Overkill on Profile Patch
**Where:** [user_profile_service.go:113-115, 153-155, 182-184](application/usecase/user_profile_service.go#L113) — `DeleteByPattern("community:threads:top:*", "community:threads:latest:*", "thread:all-communities:*")`.

**Cost:** Wipes every community feed cache on any profile change. Hot caches die.

**Fix:**
- Only invalidate communities the user is member of: look up `community_member` rows, build specific keys, delete those.
- Or: shorten TTL on community feed caches to 60-120s and stop pattern-invalidating on profile patch. Avatar/faculty stale for 1min, acceptable.

---

### 9. App-Layer Visibility Filtering
**Where:** [thread_visibility.go:10-55](application/helper/thread_visibility.go#L10-L55) — fetches all threads, builds privacy map, builds follow set, filters in Go.

**Cost:** 2-3 extra DB queries + O(n) Go loop per feed. Over-fetches before filter.

**Fix:** Push to SQL:
```sql
JOIN user_settings us ON us.user_id = threads.user_id
WHERE us.is_private = false
   OR threads.user_id = :viewer_id
   OR threads.user_id IN (
     SELECT following_id FROM user_relationships WHERE follower_id = :viewer_id
   )
```
Page limit then applies to filtered set, not pre-filter. Fewer rows shipped to Go.

---

### 10. Avatar Loads Not Cross-Endpoint Cached
**Where:** [feed_builder.go:47-81](webapi/utils/feed_builder.go#L47-L81) — batch `GetProfiles(needed)` per request.

**Fix:** `user:{id}:avatar` Redis key, TTL 1h, invalidate on profile patch. Feed builder hits Redis first, falls back to DB batch for misses only.

---

### 11. Chat Member Check Already Cached — Just Monitor
**Where:** [chat_service.go:40](application/usecase/chat_service.go#L40) — `roomRepo.IsMember` has cache `chat:member:roomID:userID`.

**Action:** No change. Watch hit rate. If <90%, bump TTL or warm cache on join/leave events.

---

## Implementation Order

1. **#1** (batch viewer flags) — biggest win, single PR, no logic risk.
2. **#2** + **#9** together — both touch `FindByRootID` and visibility filter. Push pagination + visibility into SQL in one shot.
3. **#4** (batch notification senders) — mirror of #1, easy.
4. **#3** part 1 (SQL top-N in summary) — quick.
5. **#3** part 2 + **#5** + **#6** — introduce a simple background worker pattern (Redis list + goroutine pool). Reuse for Ollama, SMTP, MinIO.
6. **#7** (cache hot reads) — once worker exists, add cache-aside.
7. **#8**, **#10** — cleanup pass.

---

## Estimated Impact

| Fix | Latency cut | DB load cut |
|---|---|---|
| #1 batch viewer flags | -70% on feed | -95% on flag queries |
| #2 paginate GetChain | -90% on deep threads | -99% rows scanned |
| #3 async summary | 30s → 50ms | -99% rows |
| #4 batch notifications | -80% | -95% |
| #5/#6 async I/O | 5-10s → 50ms | 0 |
| #7 cache reads | -90% on cache hit | -90% on hot keys |
| #9 SQL visibility | -50% on private feeds | -60% rows |
