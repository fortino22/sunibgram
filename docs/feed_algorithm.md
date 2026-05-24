# For You Feed Algorithm

`GET /threads/` returns a personalized, dynamically ordered For You feed for authenticated users. The algorithm is designed so each refresh feels alive (different order, fresh content surfacing) while pagination stays stable within a scrolling session.

## Feed Score (Stored Column)

Threads have a precomputed `feed_score` column updated every 1 minute by `infrastructure/scheduler/feed_score_scheduler.go` for content from the last 7 days:

```sql
UPDATE threads
SET feed_score = (
  like_count + 3.0 * comment_count + 5.0 * repost_count + 4.0 * bookmark_count
) / (
  1.0 + POWER(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600.0, 1.5)
)
WHERE created_at > NOW() - INTERVAL '7 days';
```

Threads older than 7 days are reset to `feed_score = 0` (only when `feed_score <> 0`, so cold rows aren't rewritten).

After the score recompute the scheduler also re-populates the **global trending cache** (`feed:trending:global`, TTL `5 min`) with the top 100 public, non-private, visibility-filtered threads ordered by `feed_score DESC`. Because the scheduler ticks every 1 min and the TTL is 5 min, the global source is effectively always cache-warm; the in-repo `forYouTrendingTTL = 2 min` only takes effect if the scheduler is down.

Combined with the read-time decay (see [Scoring](#scoring)), the effective ranking applies decay **twice**: once by the scheduler with exponent 1.5 inside `feed_score`, and once at merge time with a linear ~6h half-life. The two are intentional — the stored score determines DB-side `ORDER BY` and the read-time decay does final ranking — but be aware they compound.

## Candidate Sources

Per request the repository fetches up to 100 recent candidates from each of four sources. Each source query orders by **fresh-first then `feed_score DESC`**:

```sql
ORDER BY (threads.created_at > NOW() - INTERVAL '900 seconds') DESC,
         threads.feed_score DESC
```

Threads created inside the fresh window (`forYouFreshWindow`, 15 min) sort to the top of each source result so a freshly posted thread with `feed_score = 0` is not cut by the `LIMIT 100` before the scheduler ticks. Visibility predicates (community visibility, `is_private` author filter, follower / member joins) are untouched — the fresh ordering only changes priority, not which rows are eligible.

Filters: types `post/repost/reply`, the last 7 days. Replies are only kept when they have engagement (`like_count + repost_count + comment_count > 0`).

| Source | Filter | Source weight | Reposts |
|---|---|---|---|
| Own | viewer's own threads | `1.0` | yes |
| Global trending | public communities only, non-private authors | `1.0` | no |
| Following | threads from followed users; community-visibility checked (public OR viewer is member) | `0.8` | yes |
| Community | threads in communities the viewer is a member of | `0.7` | no (SQL fetches reposts, Go filter strips them — slight wasted work) |

The four source queries run **in parallel** via `sync.WaitGroup` inside `FindForYouFeed`. Each source query is cached in Redis (see "Caching" below) so the four queries do not hit the DB on every request. The global source is warmed by the scheduler, so under steady state only `own/following/community` ever miss cache for a given user.

## Scoring

For each candidate, an effective score is computed at merge time:

```text
age_hours = max(0, now - thread.created_at in hours)
is_fresh  = age < forYouFreshWindow (15 min)
decay     = 1 / (1 + age_hours / 6)        # log-style halflife ~6h
score     = thread.feed_score * source_weight * decay
```

Floor boosts (applied in `addCandidates`):

| Case | Floor |
|---|---|
| Own thread, fresh (age < 15 min) | `5.0` |
| Own thread, older | `0.5` |
| Non-own thread, fresh (age < 15 min) | `forYouFreshFloor` = `0.1` |
| Non-own thread, older | none — drops if `score < 0.01` |

The non-own fresh floor (`0.1`) only exists so a `feed_score = 0` post survives the `minScore = 0.01` filter; the floor value itself does not position the thread, because the post-merge fresh / ranked split sorts fresh purely by `created_at DESC`.

If the same thread appears in multiple sources, the highest score is kept (dedupe by `thread_id`). Candidates with `score < 0.01` and not in the fresh window are dropped before sorting.

## Ordering — Fresh Band + Seeded Band Shuffle

Pure score sort is deterministic and makes the feed feel frozen. Pure random shuffle destroys ranking. The algorithm splits the merged candidates into two slices, processes them independently, then concatenates `[fresh..., ranked...]`:

1. **Split** by `created_at` against `forYouFreshWindow` (15 min):
   - `fresh` — `created_at > now - 15min`
   - `ranked` — everything else
2. **Fresh band** — sort `fresh` by `created_at DESC`, then cap at `forYouFreshCap = 5`. Overflow (6th item and beyond) spills into `ranked` so the top of the feed never becomes a pure Latest tab when the network is busy.
3. **Ranked band** — sort by `score` DESC (tiebreak `created_at` DESC), then bucket candidates into **score bands** of width `0.5` (`band = floor(score / 0.5)`) and shuffle within each band using a deterministic PRNG seeded from the user's current feed seed.

The top 5 items of the feed are always the newest fresh threads (sorted by `created_at`, regardless of `feed_score`). Below that, high-scoring trending content sits in the top ranked bands, with band-internal order varying per session.

After concatenation, an **author cap of 20** is applied: at most 20 threads per author across the whole feed. For repost threads, the cap counts against the *original* author's id.

The full candidate list is then sliced for the requested page (`offset = (page-1) * limit`).

## Session Seed

A per-user 64-bit seed lives in Redis at `feed:foryou:user:{userID}:seed` with a 10-minute sliding TTL (refreshed on every read).

- First request, or after idle expiry: seed is minted from `time.Now().UnixNano() XOR userID * golden_ratio` and stored.
- Subsequent reads reuse the same seed → identical ordering → stable pagination, no duplicates across `page=1`, `page=2`, ...
- The seed is **not** invalidated by likes, comments, reposts, or new posts. Mutations only drop the source caches, so new content surfaces on the next request, but the user's current ordering is preserved.
- The seed **is** invalidated when the user explicitly refreshes via `POST /threads/feed/refresh` or after 10 minutes of feed inactivity.

This design means refreshing the page (F5) re-uses the same seed and shows the same feed — refresh-without-surprise. Pull-to-refresh / banner tap calls the refresh endpoint first, which mints a new seed and produces a visibly different order.

## Endpoints

### `GET /threads/?page=N&limit=L`

Returns one page of the For You feed. Response is the existing `PaginatedFeedResponse{ Data, Meta{ Total, Page, Limit } }`. No frontend-managed seed — the server holds it.

### `POST /threads/feed/refresh`

Auth-protected, no body, returns `204 No Content`. Deletes the caller's seed key so the next `GET /threads/` produces a freshly shuffled order. Frontend should call this on a pull-to-refresh gesture or the "Show N new posts" banner tap, then refetch page 1.

### `GET /threads/feed/new-count?since_id=<id>`

Auth-protected. Returns `{ "count": N }` — the number of feed-eligible, visibility-filtered threads with `thread_id > since_id` (strict `>`). Capped at 1000 (`LIMIT 1001` then clamped; UI should render `999+`). Result is cached per `(userID, since_id)` for 30 seconds. Frontend polls this every ~30 s to drive the "Show N new posts" banner.

**Important:** because the fresh band sorts by `created_at DESC` rather than `thread_id DESC`, two recently-created threads can land in the visible list with `id_a > id_b` even though `created_at_a < created_at_b`. The frontend must pass `Math.max(...visibleIDs)` as `since_id`, NOT `data[0].thread_id`. Passing `data[0].thread_id` undercounts new posts.

## Caching

| Key | TTL | Purpose |
|---|---|---|
| `feed:foryou:user:{userID}:seed` | 10 min (sliding) | Current feed ordering seed |
| `feed:foryou:user:{userID}:src:own` | 30 s | Own-posts candidate cache |
| `feed:foryou:user:{userID}:src:following` | 60 s | Following candidate cache |
| `feed:foryou:user:{userID}:src:community` | 60 s | Community candidate cache |
| `feed:foryou:user:{userID}:newcount:{since_id}` | 30 s | "New posts" banner counter |
| `feed:trending:global` | 5 min (scheduler-warmed every 1 min) / 2 min (lazy fallback) | Global trending candidate cache, shared across users |

Per-user keys are registered under tag `UserForYouTag(userID)` so a single `cache.InvalidateTag` drops all `src:*` and `newcount:*` entries for a user. The global key is NOT tagged — invalidation must be direct (`cache.Delete(ctx, "feed:trending:global")`) and is required when fields embedded in the cached `[]threadmodel.Thread` change (e.g. `Faculty.Update/Delete`, `Event.Update/Delete`, premium toggles for any included author — see backend `CLAUDE.md` mutation cheatsheet).

The previous per-page result cache (`feed:foryou:{userID}:{page}:{limit}`) has been removed — merge / sort / shuffle / cap run on every request but are O(800) in-memory on cached source data, so the cost is negligible compared to a freeze-the-whole-feed cache.

On thread mutations (`Save`, `Update`, `Delete`, repost, like, comment, bookmark), the helper `invalidateForYouFeedCacheForUser(ctx, userID)` deletes only that user's `src:*` and `newcount:*` keys. The seed is preserved so an active scrolling session does not visibly reshuffle when a counter ticks.

On a thread author's mutation the invalidator also runs against followers (`invalidateForYouFeedCachesForFollowers`) and community members (`invalidateForYouFeedCachesForCommunity`) so their source caches drop and they pick up the new content on the next request.

## Scaling Notes

- `feed_score` is precomputed and indexed, so per-request DB work is the 4 source queries (each capped at 100 rows). Source caches reduce DB traffic to roughly one query batch per user per minute under steady load.
- Merge, decay, sort, band-shuffle, and author-cap all run in Go on at most `forYouCandidateCap = 800` candidates.
- The new-count endpoint runs a single indexed `COUNT(*)` capped at `LIMIT 1001`, cached 30 s. Even at 30 s polling per user the DB load is bounded.
- Pagination remains offset-based for now. If active feed sessions routinely cross deeper than ~page 50, switch to cursor pagination encoding `(score, thread_id)` to avoid offset cost.
