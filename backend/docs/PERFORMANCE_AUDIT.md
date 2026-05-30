# Backend Performance / Cost Audit

Scan of the Go/Fiber backend for the most **costly** API paths — DB load, Redis round-trips,
and connection-pool starvation. Every fix below changes **only the request flow**: same business
logic, same query results, same API contract. Nothing here alters what an endpoint returns.

All `file:line` references were verified against the tree at audit time.

---

## 1. Summary

| # | Cost driver | Endpoints hit | Severity | Flow-only fix |
|---|---|---|---|---|
| P1 | DB pools 40/20; Redis pool unconfigured (default 10) | **all** | 🔴 High | Bump pool sizes; configure Redis pool |
| P2 | Per-id Redis GET/SET in a loop (O(n) RTT) | poll vote/state, user-settings privacy | 🟠 Med-High | Use existing `MGet`/`MSet` |
| P3 | `Count()` + `Find()` = 2 round-trips per page | joined/pinned lists, interactions | 🟠 Med | `COUNT(*) OVER()` window column |
| P4 | Read SELECTs on the write pool | hashtag, event, community, feedback | 🟠 Med | Switch `writeDB` → `readDB` |
| P5 | Thread media re-fetched every feed render (uncached) | all feed/list endpoints | 🟡 Med | Cache `thread:media:<id>:list` |
| P6 | Poll + options as 2 separate queries | poll fetch | 🟢 Low | `Preload("Options")` |

> **Correction to first-pass scan:** thread-media was flagged as a "CRITICAL N+1 — O(threads)
> queries per feed". **That is false.** Feed paths batch via `mediaService.GetByThreadIDs(ids)`
> ([controller_builder_helpers.go:62](../webapi/controller/_shared/controller_builder_helpers.go#L62)).
> The real issue is that the single batch query is **uncached** (P5) — medium, not critical.

---

## 2. Most costly endpoints

Ranked by per-request DB/Redis work. Drivers in brackets map to the findings below.

1. **Feed / list reads** — `GET /threads/feed/*`, `/threads/following`, `/threads/faculty`,
   `/threads/user/:id*`, `/explore/*`, `/communities/threads*`, `/likes`, `/bookmarks`.
   Each render: thread query → media batch query [P5] → avatar/meta cache. Most are already
   feed-cached, so the cost surfaces on cache-miss / refresh.
2. **Paginated joined lists** — `GET /events/joined/me`, `/events/pins/me`, `/communities/pins/me`,
   interaction listings. Two DB round-trips each [P3].
3. **Interaction mutations** — `POST/DELETE /threads/:id/like|bookmark|repost`. Trigger hashtag
   count update doing up to 3 reads, some on the write pool [P4].
4. **Poll fetch / vote** — `/threads/:id/poll*`. Per-id Redis loop [P2] + split poll/options query [P6].

Underneath all of it: connection pools [P1] cap total concurrency regardless of per-endpoint work.

---

## 3. Findings

### P1 — Connection pools starve under load · 🔴 High · ~30 min
**Where:** [config/db.go:57-61](../infrastructure/config/db.go#L57-L61) — read `MaxOpenConns=40`,
write `=20`. [cache.go:25-28](../infrastructure/cache/cache.go#L25-L28) — Redis client sets only
`Addr`/`Password`; no `PoolSize`, `MinIdleConns`, or `MaxRetries` (go-redis default pool = 10).

**Why costly:** A feed request touches the DB ≥2×. With a 40-conn read pool, sustained load past
~20 RPS queues connections → tail-latency spike. A cold Redis pool (no min-idle) churns new
connections under burst.

**Flow-only fix:** Raise `MaxOpenConns` (read ≈100, write ≈50, keep idle proportional); add
`PoolSize: 20, MinIdleConns: 5, MaxRetries: 2` to the Redis options. Ideally read from env so ops
can tune per environment. **No query, no logic change** — purely how many connections are available.

---

### P2 — In-loop Redis GET/SET instead of batch · 🟠 Med-High · ~1 h
**Where:**
- [thread_poll_repository.go:103-111](../infrastructure/repository/thread_poll/thread_poll_repository.go#L103-L111)
  — per-id `cache.Get` in a loop.
- [thread_poll_repository.go:189-196](../infrastructure/repository/thread_poll/thread_poll_repository.go#L189-L196)
  — per-id `cache.Set` + `RegisterCacheKey` in a loop.
- [user_settings_repository.go:70-81](../infrastructure/repository/user_settings/user_settings_repository.go#L70-L81)
  — per-id `cache.Get` loop.
- [user_settings_repository.go:98-100](../infrastructure/repository/user_settings/user_settings_repository.go#L98-L100)
  — per-id `cache.Set` loop.

**Why costly:** O(n) Redis round-trips per request. Fetching 100 poll/user ids = 100 sequential
RTTs where one pipelined call would do.

**Flow-only fix:** Use the already-present, already-pipelined helpers
[`Cache.MGet`](../infrastructure/cache/cache.go#L184) and
[`Cache.MSet`](../infrastructure/cache/cache.go#L208). Same keys, same values, same TTL — only
batched into one round-trip. Result set is identical, so **logic unchanged**.

---

### P3 — Count + Find = two round-trips per page · 🟠 Med · ~1-2 h
**Where:**
[event_participant_repository.go:70-85](../infrastructure/repository/event_participant/event_participant_repository.go#L70-L85),
[event_pin_repository.go:79-92](../infrastructure/repository/event_pin/event_pin_repository.go#L79-L92),
[community_pin_repository.go:79-94](../infrastructure/repository/community_pin/community_pin_repository.go#L79-L94),
[interaction_query.go:131-194](../infrastructure/repository/_shared/interaction_query.go#L131-L194).

**Why costly:** The same JOIN + WHERE runs twice — once as `Count()`, once as the paged `Find()`.
Doubles round-trips and lets the snapshot drift between the two queries.

**Flow-only fix:** Add a `COUNT(*) OVER() AS total` window column to the paged query and read `total`
off the first returned row. Same filters, same ordering, same page, same `(data, total)` contract
the service already returns — only the SQL shape changes. **No logic change.**

---

### P4 — Read SELECTs on the write pool · 🟠 Med · ~30 min
**Where:**
[hashtag_repository.go:164](../infrastructure/repository/hashtag/hashtag_repository.go#L164) & 176,
[event_repository.go:485](../infrastructure/repository/event/event_repository.go#L485), 490, 502,
[community_repository.go:326](../infrastructure/repository/community/community_repository.go#L326), 344, 363,
[feedback_repository.go:55](../infrastructure/repository/feedback/feedback_repository.go#L55).

**Why costly:** Pure SELECTs run on `r.writeDB`, loading the smaller write pool (P1) and bypassing
the read replica. The hashtag path does up to 3 such reads on every like/unlike.

**Flow-only fix:** Switch `r.writeDB` → `r.readDB` for these read-only queries.

> ⚠️ **Read-after-write caveat — apply per-site, do not blanket-swap.** Only switch sites where the
> row was **not** just written earlier in the same request. Hashtag/event/community associations are
> created at entity-creation and merely read later (on like/update), so the read replica is safe.
> [feedback_repository.go:55](../infrastructure/repository/feedback/feedback_repository.go#L55)
> re-reads a row **immediately after creating it** in the same call — under a read replica this can
> hit replication lag and return nothing. **Leave that one on `writeDB`** (or only move it if reads
> and writes share one Postgres instance). Verify each site before flipping.

---

### P5 — Thread media uncached · 🟡 Med · ~1 h
**Where:**
[thread_media_repository.go:42-49](../infrastructure/repository/thread_media/thread_media_repository.go#L42-L49)
`FindByThreadIDs` — one DB batch query on every feed render, with no Redis layer. (This is batched,
**not** N+1 — see the correction in §1.)

**Why costly:** Media for a thread is effectively immutable (only set at creation, removed at
delete) yet is re-queried from Postgres on every feed page load and every cache-miss hydration.

**Flow-only fix:** Add a cache layer keyed `thread:media:<id>:list`, read via
[`Cache.MGet`](../infrastructure/cache/cache.go#L184) for the batch path, and invalidate in the
mutation methods that already exist —
[`SaveAll`](../infrastructure/repository/thread_media/thread_media_repository.go#L28),
[`DeleteByThreadID`](../infrastructure/repository/thread_media/thread_media_repository.go#L51),
[`DeleteByIDs`](../infrastructure/repository/thread_media/thread_media_repository.go#L59). Same data
returned; cache is transparent. **No logic change.**

---

### P6 — Poll + options as two separate queries · 🟢 Low · ~30 min
**Where:**
[thread_poll_repository.go:119-136](../infrastructure/repository/thread_poll/thread_poll_repository.go#L119-L136)
— fetches polls, then fetches options in a second query and attaches them in a loop.

**Flow-only fix:** `Preload("Options", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") })`
on the poll query. One query instead of two; same attached result. **No logic change.**

---

## 4. Recommended order

1. **P1 — pools** (~30 min). Highest leverage, lowest risk: unlocks concurrency everywhere.
2. **P2 — batch Redis** (MGet/MSet already exist; mechanical).
3. **P4 — readDB swap** (per-site, honor the feedback caveat).
4. **P5 — media cache** (reuses existing cache + invalidation hooks).
5. **P3 / P6 — query merge** (window count + Preload; touch SQL, verify carefully).

---

## 5. Verification (when fixes are implemented)

- `cd backend && go vet ./...` and `go build ./...` after each change.
- `go test ./...`.
- Load-test a feed endpoint before/after **P1** to confirm the tail-latency drop.
- For **P4**, confirm the deployment's read replica vs. single-instance setup before flipping any
  read-after-write site.

---

## 6. Implementation status

| Fix | Status | Notes |
|---|---|---|
| P1 pools | ✅ Done | [db.go](../infrastructure/config/db.go) read 100/25, write 50/12; [cache.go](../infrastructure/cache/cache.go) Redis `PoolSize 20 / MinIdleConns 5 / MaxRetries 2` |
| P2 batch Redis | ✅ Done | `MGet`/`MSet` in `thread_poll` (`FindByThreadIDs`, `FindUserVotes`) + `user_settings` (`GetPrivacyMapByUserIDs`); poll-vote tag registration kept per-key |
| P4 readDB swap | ✅ Done (scoped) | Switched hot reads in `hashtag` (`PatchHashtagCounterByThreadID`) + `event` (`PatchEventCounterByThreadID`). **Left on writeDB:** `community.Delete/Deactivate/Reactivate` (rare, replica-lag risk) and `feedback:55` (read-after-create) |
| P5 media cache | ✅ Done | `thread:media:<id>:list` (24h), `MGet` batch, empty-slice cached, invalidated in `SaveAll`/`DeleteByThreadID`/`DeleteByIDs` |
| P6 poll Preload | ✅ Done | Batch poll path now single query with `Preload("Options")` |
| **P3 Count+Find** | ⛔ Deferred | `COUNT(*) OVER()` loses total on empty/last page → behavior change. Sites already cache-fronted (low value). Left as-is. |

Verified: `go build ./...`, `go vet ./...`, `go test ./...` all clean.
