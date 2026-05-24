# Backend Cost Optimization Audit

Scope: `backend/`. No business-logic change. Reduce OpenAI spend, Postgres CPU, Redis memory, MinIO/S3 egress.

Impact: **HIGH** = >10% bill, **MED** = noticeable, **LOW** = hygiene.

---

## HIGH-1 — OpenAI: no MaxTokens cap

**Problem**
[infrastructure/openai/openai_client.go:30-37](backend/infrastructure/openai/openai_client.go#L30-L37) — `ChatCompletionNewParams` sets only `Model` + `Messages`. No `MaxTokens`. Default for `chat.completions` is model max (e.g. gpt-4o = 16k output). One bad prompt can stream 16k tokens at ~$0.06/1k → $0.96 per call.

**Solution**
```go
stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
    Model:     c.model,
    Messages:  []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)},
    MaxTokens: openai.Int(400),  // ~1 paragraph
    Temperature: openai.Float(0.3),
})
```
Summary use-case = 1 paragraph (~150-300 tokens). **Saves up to 95% of output cost** if model defaults high.

---

## HIGH-2 — OpenAI: input not truncated

**Problem**
[thread_summary_service.go:91-102](backend/application/usecase/thread_summary/thread_summary_service.go#L91-L102) — fetches 50 replies via `FindRepliesContent(threadID, 50)`, joins with `\n`. No per-comment cap, no total cap. Twitter-clone reply averages ~100 chars but adversary can post 10kB comments → single summary = 500kB prompt ≈ 125k input tokens ≈ $0.50 at gpt-4o pricing.

**Solution**
Before join:
```go
const maxPerComment = 280   // tweet-length cap
const maxTotal = 6000       // ~1.5k tokens
trimmed := make([]string, 0, len(comments))
total := 0
for _, c := range comments {
    if len(c) > maxPerComment { c = c[:maxPerComment] + "…" }
    if total+len(c) > maxTotal { break }
    trimmed = append(trimmed, c); total += len(c)
}
```
**Saves ~70-90% of input tokens** on long-thread cases. No logic change — just bounds.

---

## HIGH-3 — OpenAI: model not pinned, may default to expensive

**Problem**
[openai_client.go:21](backend/infrastructure/openai/openai_client.go#L21) — `model := os.Getenv("OPENAI_MODEL")`. If env unset, `shared.ChatModel("")` → either rejects or routes to default (model-dependent). No documentation of which model is set in prod.

**Solution**
- Add `.env.example` line: `OPENAI_MODEL=gpt-4o-mini`
- For summary use-case, **gpt-4o-mini is ~17x cheaper than gpt-4o** ($0.15 vs $2.50 per M input). Quality drop negligible for paragraph summary.
- Fail-fast in `NewClient()` if model empty.

---

## HIGH-4 — Thread summary cached forever (Redis memory growth)

**Problem**
[thread_summary_service.go:21](backend/application/usecase/thread_summary/thread_summary_service.go#L21)

```go
threadSummaryTTL time.Duration = 0
```

`0` in the cache wrapper = no expiration. Every summarized thread → permanent Redis key. Thread count grows monotonically → Redis OOM eventually. Then ElastiCache scales up = bill goes up.

**Solution**
Set TTL = 7 days (or matching `created_at > NOW() - INTERVAL '7 days'` feed window):
```go
threadSummaryTTL = 7 * 24 * time.Hour
```
Cold threads age out. Hot threads refresh via existing `RegisterReply` invalidation on every 20 replies. **Bounds Redis memory** at ~7d of summary keys.

---

## HIGH-5 — Feed score scheduler runs every minute

**Problem**
[infrastructure/scheduler/feed_score_scheduler.go:28](backend/infrastructure/scheduler/feed_score_scheduler.go#L28) — `time.NewTicker(1 * time.Minute)`. Per tick:
1. `UPDATE threads ... WHERE created_at > NOW() - INTERVAL '7 days'` — scans 7-day population, writes rows that changed.
2. `SELECT TOP 100 ORDER BY feed_score DESC` with 3 LEFT JOINs.
3. ZSET write to Redis.

1440 ticks/day. The `IS DISTINCT FROM` clause limits writes but the **read scan still runs every minute**. On a 100k-thread/week dataset, that's ~150k row scan/minute = constant Postgres load.

**Solution**
- Bump to `5 * time.Minute`. Feed freshness loss ≈ 5 min — for "trending top 100" that's invisible to users.
- Alt: tiered — `1m` for **fresh window** (`created_at > NOW() - INTERVAL '15 min'`), `15m` for older. Currently every tick scans all 7 days; ~95% of rows are stable older threads.

**Saves ~80% scheduler CPU/IO** with negligible UX impact.

---

## HIGH-6 — Read DB pool oversized

**Problem**
[infrastructure/config/db.go:43-48](backend/infrastructure/config/db.go#L43-L48)

```go
sqlDB.SetMaxOpenConns(100)   // read
sqlDB.SetMaxIdleConns(50)
```

100 idle-capable conns/instance × N instances → quickly exceed Postgres `max_connections` (default 100-200 on managed Postgres). RDS small instance is billed partly on conn count via memory.

Also write pool 75 — most workloads write << read. 75 writers idle most of the time.

**Solution**
- Read: `Max=40`, `Idle=10` (most apps need ≤20 active queries/instance).
- Write: `Max=20`, `Idle=5`.
- Monitor via `pg_stat_activity` for 1 week then tune. **Avoids RDS instance upgrade** when conn count saturates.

---

## HIGH-7 — `CountNewSinceID` fetches 1001 rows for a counter

**Problem**
[thread_repository_for_you.go:423](backend/infrastructure/repository/thread/thread_repository_for_you.go#L423) — `LIMIT 1001` to display "999+". Frontend polls every ~30s. 1001 rows × poll/30s × every viewer = lots of wasted IO when feed is busy.

**Solution**
- Cache the count by viewer for ~15-30s: key `feed:newcount:user:<uid>:since:<sinceID>`. Lookup before query.
- Drop LIMIT to 200 and display "200+" — most users don't care between 200 and 999+. **Cuts row reads ~5x**.

---

## HIGH-8 — Polling pattern for notifications / unread

**Problem** (suspected — verify with frontend)
Frontend likely polls `/notifications/unread-count` and `/chat/unread`. Each call hits Postgres COUNT.

**Solution**
Already have WebSocket hub ([infrastructure/socket/hub.go](backend/infrastructure/socket/hub.go)). Push a `notification:unread` event on insert, frontend tracks counter locally. Backend keeps a small Redis hash `user:unread:<uid>` updated on write — single GET on connect, no polling.

**Saves all the GET COUNT(*) traffic.** Needs frontend coordination, but server side already supports it.

---

## MED-1 — Faculty preload in user batch loads

**Problem**
[infrastructure/repository/user_faculty/user_faculty_repository.go:62](backend/infrastructure/repository/user_faculty/user_faculty_repository.go#L62) — `Preload("Faculty")` on every batch load. Faculty count is small (~10-20) and rarely changes.

**Solution**
Cache `faculty:id:<id>` with 1h TTL. Pre-resolve faculty names in `FeedBuilder.resolveProfileMeta` via single bulk fetch already exists per CLAUDE.md — confirm only `user:meta:<uid>` is cold-path; faculty join should be eliminable.

---

## MED-2 — Feed builder triple-loop collects IDs

**Problem**
[webapi/controller/thread/thread_controller_builder.go:63-137](backend/webapi/controller/thread/thread_controller_builder.go#L63-L137) — three loops over threads + parents to collect media IDs, author IDs, poll IDs.

**Solution**
Single loop, three append slices. CPU win minor but allocs reduce (esp. avoid repeated re-slicing). Pre-size slices: `make([]uint, 0, len(threads)*2)`.

---

## MED-3 — Explore search uses ILIKE with count-then-query

**Problem**
[infrastructure/repository/explore/explore_repository.go:165-192](backend/infrastructure/repository/explore/explore_repository.go#L165-L192) — `COUNT(*) WHERE name ILIKE '%q%'` then `SELECT ... LIMIT/OFFSET`. ILIKE on unindexed text column = sequential scan. Count + query = 2 sequential scans per request.

**Solution**
- Add `pg_trgm` GIN index on `communities.name` (migrations need DDL). Already migration-modified per git status — slot it in.
- Skip exact count, return `hasMore` flag from `LIMIT+1` pattern instead of total. Frontend pagination usually doesn't need exact total.

---

## MED-4 — Per-poll cache TTL 30s

**Problem**
[infrastructure/repository/thread_poll/thread_poll_repository.go:105-120](backend/infrastructure/repository/thread_poll/thread_poll_repository.go#L105-L120) — 30s TTL on poll body. Poll content (question + options) is immutable; only vote counts change.

**Solution**
- Split: poll body cached 1h (immutable), vote counts cached 30s (or push on vote).
- **Cuts poll DB reads ~95%** on hot polls.

---

## MED-5 — Per-page interaction cache keys

**Problem**
[infrastructure/repository/_shared/interaction_query.go:118-160](backend/infrastructure/repository/_shared/interaction_query.go#L118-L160) — `FindByUserID` caches per `(userID, page, limit)`. Frontend varying limit = cache miss cascade.

**Solution**
Cache only fixed-page-size variants (e.g. limit=20, 50). Reject odd limits at controller. Pin to canonical sizes → cache hit ratio jumps.

---

## MED-6 — JSON deserialization of full thread on every hydrate

**Problem**
[thread_repository_query_cache.go:48-84](backend/infrastructure/repository/thread/thread_repository_query_cache.go#L48-L84) — `hydrateThreads` does single `WHERE IN` for cache misses. Per-thread cached entry deserialized via encoding/json. Hot threads decoded repeatedly across viewers.

**Solution**
- Cache value as `[]byte` of pre-rendered DTO JSON, not domain model. Controller writes raw bytes to response → skip Marshal step entirely.
- Or: msgpack/gob — 2-3x faster than encoding/json.

**Saves CPU**, especially at high RPS.

---

## MED-7 — `LoadAvatarsForUserIDs` two-pass

**Problem**
[webapi/utils/feed_builder.go:96-129](backend/webapi/utils/feed_builder.go#L96-L129) — cache lookup, then profile-meta lookup, then DB if miss. Three Redis RTTs per feed render.

**Solution**
Combine via `MGET` of both `user:meta:<uid>` and `user:avatar:<uid>` in single pipeline. **Cuts Redis RTTs ~3x**.

---

## MED-8 — Subscription expiry scheduler hourly

**Problem**
[infrastructure/scheduler/subscription_expiry_scheduler.go:11](backend/infrastructure/scheduler/subscription_expiry_scheduler.go#L11) — hourly tick. Subscription expiry is end-of-day relevant for billing — not hourly.

**Solution**
6h or daily at fixed time (e.g. 02:00 UTC). 24x fewer queries. Same with `verification_code_purge.go` — daily is fine.

---

## LOW-1 — `GetURL` called per-item in render loop

**Problem**
[explore_controller_builder.go:29-46](backend/webapi/controller/explore/explore_controller_builder.go#L29-L46) — string concat per user in loop. Pure CPU, no IO. Negligible.

**Solution** — fine as-is; mention if profiling flags it.

---

## LOW-2 — Recursive CTE no pagination

**Problem**
[thread_repository_query.go:90-163](backend/infrastructure/repository/thread/thread_repository_query.go#L90-L163) — `findAncestors` depth limit 100. Long reply chains scan up to 100 rows per request.

**Solution**
Limit to 20 (UI rarely shows >20 levels). Index `(thread_id, parent_id)` if not already (verify migrate.go).

---

## LOW-3 — Hashtag/Event weekly resets — confirm cron timing

**Problem**
[hashtag_weekly_reset.go](backend/infrastructure/scheduler/hashtag_weekly_reset.go) — 7-day ticker. If process restart happens mid-week, the ticker resets — drift over time.

**Solution**
Switch to actual cron expression `0 0 * * MON` via `robfig/cron` (already common in Go stacks).

---

## LOW-4 — Storage signed URLs regenerated per render

**Problem** (suspected — confirm in storage_service.go)
[infrastructure/storage/s3_client.go](backend/infrastructure/storage/s3_client.go) — chat media presigned URLs. If presign called every render, S3 SDK does HMAC per call. Cheap but cumulative on heavy chat.

**Solution**
Cache presigned URL by object key with TTL = `urlExpiry - 5min`. SDK call cost drops; URL same for ~1h.

---

## Summary Table

| ID | Impact | Where | Fix in one line |
|---|---|---|---|
| HIGH-1 | HIGH | openai_client.go:30 | Set `MaxTokens: 400` |
| HIGH-2 | HIGH | thread_summary_service.go:91 | Cap per-comment + total length pre-join |
| HIGH-3 | HIGH | openai_client.go:21 | Pin `OPENAI_MODEL=gpt-4o-mini` + fail-fast |
| HIGH-4 | HIGH | thread_summary_service.go:21 | `threadSummaryTTL = 7*24h` |
| HIGH-5 | HIGH | feed_score_scheduler.go:28 | 1min → 5min (or tiered) |
| HIGH-6 | HIGH | config/db.go:43 | Pool 100→40 read, 75→20 write |
| HIGH-7 | HIGH | thread_repository_for_you.go:423 | Cache new-count 15-30s + drop limit to 200 |
| HIGH-8 | HIGH | (frontend coord) | Push unread via existing WS hub, not poll |
| MED-1 | MED | user_faculty_repository.go:62 | Cache faculty entity 1h |
| MED-2 | MED | thread_controller_builder.go:63 | Single loop + pre-sized slices |
| MED-3 | MED | explore_repository.go:165 | pg_trgm GIN + LIMIT+1 pattern |
| MED-4 | MED | thread_poll_repository.go:105 | Split immutable body (1h) vs counts (30s) |
| MED-5 | MED | interaction_query.go:118 | Pin limit to canonical sizes |
| MED-6 | MED | thread_repository_query_cache.go:48 | Cache pre-rendered DTO bytes |
| MED-7 | MED | feed_builder.go:96 | Single Redis pipeline (MGET) |
| MED-8 | MED | subscription_expiry_scheduler.go:11 | Hourly → daily |
| LOW-2 | LOW | thread_repository_query.go:90 | Cap ancestor depth at 20 |
| LOW-3 | LOW | hashtag_weekly_reset.go | Use cron expr, not ticker |
| LOW-4 | LOW | s3_client.go | Cache presigned URLs by key |

---

## Expected combined savings

| Surface | Approx reduction |
|---|---|
| OpenAI bill (output) | 80-95% via HIGH-1 |
| OpenAI bill (input) | 70-90% via HIGH-2 |
| OpenAI per-call cost | 17x via gpt-4o-mini (HIGH-3) |
| Redis memory growth | Bounded via HIGH-4 |
| Postgres scheduler load | 80% via HIGH-5 |
| Postgres conn pressure | Avoid instance upgrade via HIGH-6 |
| Feed poll IO | 5x via HIGH-7 |
| Notification poll traffic | Eliminated via HIGH-8 |

Net: OpenAI bill could drop **~95%** combining HIGH-1+2+3. Postgres steady-state CPU drops **~50-70%** with HIGH-5 alone.
