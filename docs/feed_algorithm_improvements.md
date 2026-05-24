# For You Feed — Improvement Backlog

Scope: ranking quality, performance, correctness, and scaling for `FindForYouFeed` in [`backend/infrastructure/repository/thread/thread_repository_for_you.go`](../backend/infrastructure/repository/thread/thread_repository_for_you.go) and the scheduler in [`backend/infrastructure/scheduler/feed_score_scheduler.go`](../backend/infrastructure/scheduler/feed_score_scheduler.go).

Each item is independently shippable. Ordered roughly by ROI (impact per implementation effort) within each section.

---

## 1. Performance — DB & Query

### 1.1 Required composite indexes (verify and add if missing)

Every source query has the shape:

```sql
WHERE created_at > NOW() - INTERVAL '7 days'
  AND type IN (...)
  AND <source-specific predicate>
ORDER BY (created_at > NOW() - INTERVAL '900 seconds') DESC,
         feed_score DESC
LIMIT 100;
```

A naive `(created_at)` or `(feed_score)` index forces a sort. Recommended:

| Index | Sources served | Notes |
|---|---|---|
| `(user_id, created_at DESC) WHERE deleted_at IS NULL` | `findForYouOwn` | own feed; also reused by profile feed |
| `(community_id, created_at DESC) WHERE community_id IS NOT NULL` | `findForYouCommunity` | composite covers the JOIN |
| `(feed_score DESC, created_at DESC) WHERE created_at > NOW() - INTERVAL '7 days'` (partial) | `findForYouGlobal` | partial index keeps it small; the 7-day predicate makes Postgres use it |
| `(created_at DESC, thread_id DESC)` | new-count, sort tiebreak | also useful for explore feed |

Run `EXPLAIN (ANALYZE, BUFFERS)` on each source query in production-like data before deciding. The fresh-first `ORDER BY` confounds index-only scans — see 1.3.

### 1.2 The fresh-first `ORDER BY` blocks index-driven sort

```sql
ORDER BY (created_at > NOW() - INTERVAL '900 seconds') DESC, feed_score DESC
```

The boolean expression is not in any composite index, so Postgres always sorts in memory. With 100-row LIMIT and source filters, this is cheap today, but it scales poorly.

**Alternative:** issue two cheaper queries per source and union in Go:

```sql
-- 1) Fresh window — at most a handful of rows
SELECT ... WHERE created_at > NOW() - INTERVAL '900 seconds' AND <filter>
ORDER BY created_at DESC LIMIT 100;

-- 2) Score-ranked — uses (feed_score DESC) index
SELECT ... WHERE created_at <= NOW() - INTERVAL '900 seconds' AND <filter>
ORDER BY feed_score DESC LIMIT 100;
```

Merge the two slices in Go; total candidates per source still ≤ 100 after dedupe. Lets the planner use indexes for both halves.

### 1.3 Replace `NOW() - INTERVAL` with bound parameters

`NOW()` inside a query is volatile, which makes Postgres skip plan caching and re-plan every call. Bind the cutoff timestamp in Go:

```go
cutoff := time.Now().Add(-7 * 24 * time.Hour)
freshCutoff := time.Now().Add(-forYouFreshWindow)
query.Where("threads.created_at > ?", cutoff)
     .Order("(threads.created_at > ?) DESC, threads.feed_score DESC", freshCutoff)
```

Side benefit: identical SQL across requests → pgbouncer/prepared-statement cache hits.

### 1.4 `PreloadThreadDetails` over-fetches per source

Each source query preloads the full author + community + parent + events graph for up to 100 rows. After merge+dedupe+author-cap, ~80% of those rows are discarded. Two options:

- **A. Defer the heavy preload to after pagination.** Source queries fetch only the columns needed for ranking (`thread_id`, `user_id`, `community_id`, `type`, `created_at`, `feed_score`, counters). After `applyAuthorCap` slices to the page, hydrate the 20-or-so visible rows with `PreloadThreadDetails` in one round trip.
- **B. Keep current shape but trim preloads.** `OriginalThread.ThreadEvents.Event` is rarely used at feed-list time — verify the DTO actually reads them before paying the cost.

Option A reduces per-request DB bytes by ~5–10×.

### 1.5 `findForYouGlobal` duplicates the scheduler's work

The global trending list is rebuilt every minute by the scheduler with TTL 5 min. The repo function still has a lazy-load path that re-runs the same query on cache miss. In normal operation that path is dead code; under scheduler downtime it stampedes (every user request runs the same global query simultaneously).

**Fix:** lazy path should use `singleflight.Group` keyed by `forYouSourceGlobalKey`, or simply return an empty slice and log — the scheduler will re-warm within 60 s.

### 1.6 Cache stampede on per-user source keys

`findForYouOwn/Following/Community` have `Get → on miss, DB → Set`. On TTL expiry, every concurrent request from the same user re-runs the DB query. Wrap with `singleflight.Group` keyed by the cache key; coalesces concurrent misses to one DB hit.

### 1.7 Scheduler `UPDATE threads SET feed_score = ...` is a full table write every minute

Even with the 7-day predicate, this rewrites every row matching the predicate every minute. On a busy table this generates churn (WAL, autovacuum pressure, replica lag).

Options, smallest blast-radius first:

- **A. Skip rows that haven't moved.** Add `AND feed_score IS DISTINCT FROM <new-value>` — Postgres still computes both sides but skips the row update + WAL.
- **B. Tighten the recompute window for old rows.** Threads aged > 24h barely change feed_score per minute (decay is sub-percent). Recompute > 24h rows every 10 min, < 24h every minute.
- **C. Move recompute to event-driven.** Recompute a thread's `feed_score` inside the same transaction that increments a counter. Drop the scheduler entirely. Higher correctness, more code churn.

(C) is the long-term right answer but (A) is one line.

### 1.8 Author-cap is O(n) per request but allocates per row

`applyAuthorCap` builds a fresh `map[uint]int` of size `len(candidates)` (up to 800) on every request. Negligible at current scale, but a `[800]uint{}` arena pool would eliminate the allocs.

---

## 2. Performance — Caching

### 2.1 The four source caches are independent

A `like` on a thread invalidates `src:own/following/community` for the author + followers + community members. The current code uses `cache.InvalidateTag(UserForYouTag(uid))` — good. But under churn, the source caches expire faster than they're useful.

**Fix:** decouple read-time freshness from invalidation. Set TTL to 5 min uniformly, rely on tag invalidation for correctness. Current 30 s / 60 s TTLs were chosen before tag-invalidation existed.

### 2.2 `feed:trending:global` is one big Redis value

Top-100 threads with full preload graph ≈ 100–500 KB in JSON. Every request deserializes it. Two ideas:

- Store it as a Redis sorted set of `thread_id` with score = `feed_score`, then `MGET thread:id:<id>` for hydration (already cached by `ThreadFeedsTag`).
- Compress with `gzip`/`zstd` if you keep the monolithic shape (50–80% size reduction; Redis CPU is cheap).

### 2.3 Negative caching for users with no follows / no communities

A new user's `findForYouFollowing` runs the JOIN, gets zero rows, caches `[]`. That's already correct, but the query still hits the DB once per cache miss. If `count(user_relationships WHERE follower_id = uid)` is cached, skip the source query entirely.

### 2.4 Newcount cache key explodes with `since_id`

`feed:foryou:user:{userID}:newcount:{since_id}` — every distinct `since_id` produces a new key. With pagination + scrolling, a single user generates dozens of keys per session. They're all tagged so invalidation is fine, but Redis memory pressure grows.

**Fix:** floor `since_id` to a bucket (e.g. `since_id / 100 * 100`) before keying. Same user polling with `since_id = 12345` and `12346` should hit the same cache entry (the count differs by at most a handful). Or just stop caching this — the query is `COUNT(*)` with `LIMIT 1001`, ~1 ms with a `(created_at DESC, thread_id)` index.

---

## 3. Ranking & Algorithm

### 3.1 Double decay is masking the actual ranking

Stored decay: `1 / (1 + age_h^1.5)` (scheduler)
Read decay: `1 / (1 + age_h / 6)` (merge)
Effective: `raw_interactions / ((1 + age_h^1.5) * (1 + age_h / 6))`

At `age_h = 6` (half-life intent): effective decay = `1 / (1 + 14.7) / (1 + 1) = 1/29.4` — far steeper than the documented "6h half-life". Recommend picking one:

- **Option A — keep stored decay only.** Remove the read-time `decay` factor; rely on `feed_score` directly. Lose per-request freshness response (the read-time decay updates between scheduler ticks); gain interpretability.
- **Option B — keep read decay only.** Store raw weighted interactions in `feed_score`. Apply all decay at read time. More CPU per request, far more predictable.
- **Option C — explicit two-stage.** Document the compound decay as intentional and rename one of them (e.g. `stored_decay_exp = 1.5`, `freshness_decay_halflife = 6h`).

Without picking, ranking is hard to reason about.

### 3.2 Source weights are static and untested

`{own: 1.0, global: 1.0, following: 0.8, community: 0.7}` — these were picked by intuition. With no A/B harness, you can't tell if `following: 1.2` would improve session length.

**Foundation:** instrument feed impressions + engagements (`thread_id`, `position`, `source`, `was_clicked`, `was_liked`). Even without an experiment framework, raw logs let you compute per-source CTR and decide weights from data.

### 3.3 Author cap of 20 is too high for top-of-feed concentration

Counter-intuitive but: a user followed by the viewer who posts 20 times today can dominate position 1–20 in the ranked band (after the fresh cap). Two fixes, applied together:

- **Decreasing cap by position.** Allow ≤ 3 from one author in positions 1–10, ≤ 5 in 1–25, ≤ 20 overall.
- **Per-author score penalty after first 3.** Each additional candidate from the same author gets `score *= 0.7^k` where k = rank within author.

This is what production feeds (Twitter, Threads, IG) do.

### 3.4 No exploration / no negative feedback

The feed never deliberately surfaces an unseen author or topic, and it has no concept of "viewer hid / blocked / muted this thread". Two cheap wins:

- **ε-greedy injection.** With probability 0.05 per request, replace one ranked-band slot with a random eligible thread the viewer hasn't seen. Cheap; expands the candidate set the system can learn from.
- **Hide signal.** Add `thread_hides` table + `POST /threads/{id}/hide`. Filter hidden threads at SQL level in all four sources. Massive UX win; trivial implementation.

### 3.5 The fresh band is unbounded by total feed share

Doc: "top 5 items always newest." With a slow community, all 5 fresh slots may be filled by low-engagement noise from the viewer's own community while top global content sits below. The 5 hard slots are too generous.

**Tweak:** `fresh_cap = min(5, ceil(0.2 * page_limit))`. With `limit = 20`, that's 4 slots. With `limit = 10` (mobile), 2 slots. Keeps fresh share proportional.

### 3.6 Replies are bucketed identically to posts

A reply with `like_count = 1` has the same scoring path as a top-level post with `like_count = 1`. But replies have different intent (often conversational, less discovery-worthy). Apply a `reply_weight = 0.5` multiplier so they don't dominate `following` source for chatty users.

### 3.7 No personalization signal beyond follow graph

Current personalization = `following` and `community` source weight. Nothing about:
- Authors the user *likes/comments on* but doesn't follow
- Hashtags the user engages with
- Time-of-day patterns

Foundation work: store last 100 interactions per user (Redis list or PG table). Compute a "warm authors" set (authors viewer interacted with in last 14 days). Promote their threads with a `warm_author_weight` multiplier in `addCandidates`. This is the single biggest ranking-quality lever.

### 3.8 The 7-day window is hard-coded everywhere

Five places hard-code `'7 days'`. Pull into a constant `forYouRecencyWindow = 7 * 24 * time.Hour`. Trivial; reduces drift risk if you ever tune.

### 3.9 `feed_score = 0` reset is unnecessary

`UPDATE threads SET feed_score = 0 WHERE created_at <= NOW() - INTERVAL '7 days' AND feed_score <> 0` runs every minute. After the first run it's idempotent. The source queries already filter `created_at > NOW() - INTERVAL '7 days'`, so cold rows never enter the candidate pool. Drop the reset query; saves a table scan per minute.

---

## 4. Correctness & Observability

### 4.1 Seed mint can theoretically still produce zero

```go
seed = time.Now().UnixNano() ^ int64(uint64(userID)*0x9E3779B97F4A7C15)
if seed == 0 { seed = 1 }
```

The `if seed == 0` branch handles the collision, but it produces a single-seed degenerate ordering for all users that hit it. Use `binary.LittleEndian.Uint64` from `crypto/rand` if you want true uniqueness; the current XOR is fine in practice but logs of seed = 1 would be useful to spot the case.

### 4.2 No metric on cache hit / miss / source latency

Wrap each source function with metrics:
- `feed.foryou.source.{name}.cache_hit` (counter)
- `feed.foryou.source.{name}.db_latency_ms` (histogram)
- `feed.foryou.merge_total_candidates` (histogram)
- `feed.foryou.author_cap_drops` (counter — how many candidates the author cap removes)

Without these, ranking changes are blind.

### 4.3 `attachVisibleParents` runs after the author cap

The page slice is already finalized when `attachVisibleParents` runs. If a parent is hidden (private author / blocked community), the child reply is still in the result but with a missing parent. Verify the controller / DTO handles this — otherwise the UI renders an orphan reply.

### 4.4 `total` is computed before pagination but after author-cap

`total = len(capped)`. With offset-based pagination, `total` can shrink between page 1 and page 2 if `feed_score` changes or new posts arrive — frontend may show "page 3 of 47" then "page 3 of 39". This is acceptable for an infinite-scroll UI; problematic for any UI showing total count. Document the semantics.

### 4.5 New-count strict `>` vs fresh-band sort by `created_at`

Already noted in the algorithm doc. Add a backend log line when `since_id < max_visible_id` to confirm clients are passing the correct value; old clients passing `data[0].thread_id` would surface as a continuous undercount.

### 4.6 Scheduler error handling drops on first failure

```go
if err := s.db.Exec(...).Error; err != nil {
    log.Printf("feed score recompute failed: %v", err)
    return
}
```

The `return` skips the global cache refresh entirely. A flaky `feed_score` UPDATE leaves the global trending cache untouched until the next minute — usually fine, but the failure is invisible past the log line. Push to a metric (`feed.scheduler.recompute_failures`) and alert.

---

## 5. Scaling Roadmap

### 5.1 Cursor pagination once active sessions cross page ~30

Offset-based pagination is `O(page * limit)` on the candidate list. At `page = 50, limit = 20`, you build 1000 candidates and throw away 980. Encode the cursor as `(score, thread_id)` and resume from there. Mentioned in current doc — actually implement before it hurts.

### 5.2 Pre-compute candidate set per user

For users with > 10 active sessions / day, a warm cache of merged candidates (post fresh-band + ranked-band shuffle, pre-author-cap) keyed by `(userID, seed)` would let `page=2..N` skip the merge entirely. Cache TTL = seed TTL (10 min). Memory cost: ~50 KB per active user.

### 5.3 Move fanout invalidation off the request path

`invalidateForYouFeedCachesForFollowers` and `invalidateForYouFeedCachesForCommunity` run synchronously on every author mutation. For a user with 10k followers this iterates 10k cache deletes. Push the fanout into a Redis Stream / NATS queue; a worker drains it. Mutations return in O(1).

### 5.4 Read replicas

`r.db.readDB` already separates write/read paths. Verify all four source queries actually land on a replica under load (pgbouncer config, replication lag). Source caches mean replica reads should be rare per user per minute, but during cache-cold periods (deploys, Redis restart) every user falls through to DB.

---

## Quick-win shortlist (≤ 1 day each)

1. **3.9** — Drop the `feed_score = 0` reset query (one line).
2. **3.8** — Extract `forYouRecencyWindow` constant.
3. **1.3** — Bind `cutoff` / `freshCutoff` as parameters, not `NOW()`.
4. **1.7 A** — Add `IS DISTINCT FROM` to scheduler UPDATE.
5. **4.2** — Wire cache hit/miss + DB latency metrics.
6. **3.5** — Make `fresh_cap` proportional to `page_limit`.
7. **1.6** — `singleflight` around the three per-user source caches.
8. **2.4** — Floor `since_id` to a bucket OR stop caching newcount.

## Mid-term (1–3 days each)

- **1.1 / 1.2** — Index audit + fresh-vs-score split queries.
- **1.4** — Two-stage hydration (rank cheaply, hydrate page only).
- **3.3** — Position-aware author cap + per-author penalty.
- **3.4** — Hide signal + ε-greedy exploration.
- **4.2 + 3.2** — Impression logging + per-source CTR analysis.

## Long-term (> 1 week)

- **3.7** — Warm-author personalization.
- **5.1** — Cursor pagination.
- **5.3** — Async fanout invalidation.
- **1.7 C** — Event-driven `feed_score` (replaces scheduler).
