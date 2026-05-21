# For You Feed Algorithm

`GET /threads/` returns a personalized, dynamically ordered For You feed for authenticated users. The algorithm is designed so each refresh feels alive (different order, fresh content surfacing) while pagination stays stable within a scrolling session.

## Feed Score (Stored Column)

Threads have a precomputed `feed_score` column updated every 5 minutes by `infrastructure/scheduler/feed_score_scheduler.go` for content from the last 7 days:

```text
(likes + 3 * comments + 5 * reposts + 4 * bookmarks)
/ (1 + (age_hours ^ 1.5))
```

Threads older than 7 days are reset to `feed_score = 0`.

## Candidate Sources

Per request the repository fetches up to 100 recent candidates from each of four sources, ordered by `feed_score DESC`, filtered to types `post/repost/reply` and the last 7 days. Replies are only kept when they have engagement (`like_count + repost_count + comment_count > 0`).

| Source | Filter | Source weight | Reposts |
|---|---|---|---|
| Own | viewer's own threads | `1.0` | yes |
| Global trending | public, non-private authors | `1.0` | no |
| Following | threads from followed users; community-visibility checked | `0.8` | yes |
| Community | threads in communities the viewer is a member of | `0.7` | no |

Each source query is cached in Redis (see "Caching" below) so the four queries do not hit the DB on every request.

## Scoring

For each candidate, an effective score is computed at merge time:

```text
age_hours = max(0, now - thread.created_at in hours)
decay     = 1 / (1 + age_hours / 6)        # log-style halflife ~6h
score     = thread.feed_score * source_weight * decay
```

If the same thread appears in multiple sources, the highest score is kept (dedupe by `thread_id`). Candidates with `score < 0.01` are dropped before sorting.

## Ordering — Seeded Band Shuffle

Pure score sort is deterministic and makes the feed feel frozen. Pure random shuffle destroys ranking. The algorithm uses a two-step ordering:

1. Sort all surviving candidates by `score` DESC (tiebreak `created_at` DESC).
2. Bucket candidates into **score bands** of width `0.5` (`band = floor(score / 0.5)`), then shuffle the candidates **within each band** using a deterministic PRNG seeded from the user's current feed seed.

High-scoring (trending) content always sits in the top bands, but the exact order inside each band varies between sessions, so the feed looks different on every refresh without ever moving a low-score post above a high-score post.

After shuffling, an **author cap of 20** is applied: at most 20 threads per author across the whole feed. For repost threads, the cap counts against the *original* author's id.

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

Auth-protected. Returns `{ "count": N }` — the number of feed-eligible, visibility-filtered threads with `thread_id > since_id`. Capped at 1000 (UI should render `999+`). Result is cached per `(userID, since_id)` for 30 seconds. Frontend polls this every ~30 s to drive the "Show N new posts" banner.

## Caching

| Key | TTL | Purpose |
|---|---|---|
| `feed:foryou:user:{userID}:seed` | 10 min (sliding) | Current feed ordering seed |
| `feed:foryou:user:{userID}:src:own` | 30 s | Own-posts candidate cache |
| `feed:foryou:user:{userID}:src:following` | 60 s | Following candidate cache |
| `feed:foryou:user:{userID}:src:community` | 60 s | Community candidate cache |
| `feed:foryou:user:{userID}:newcount:{since_id}` | 30 s | "New posts" banner counter |
| `feed:trending:global` | 2 min | Global trending candidate cache |

The previous per-page result cache (`feed:foryou:{userID}:{page}:{limit}`) has been removed — merge / sort / shuffle / cap run on every request but are O(800) in-memory on cached source data, so the cost is negligible compared to a freeze-the-whole-feed cache.

On thread mutations (`Save`, `Update`, `Delete`, repost, like, comment, bookmark), the helper `invalidateForYouFeedCacheForUser(ctx, userID)` deletes only that user's `src:*` and `newcount:*` keys. The seed is preserved so an active scrolling session does not visibly reshuffle when a counter ticks.

On a thread author's mutation the invalidator also runs against followers (`invalidateForYouFeedCachesForFollowers`) and community members (`invalidateForYouFeedCachesForCommunity`) so their source caches drop and they pick up the new content on the next request.

## Scaling Notes

- `feed_score` is precomputed and indexed, so per-request DB work is the 4 source queries (each capped at 100 rows). Source caches reduce DB traffic to roughly one query batch per user per minute under steady load.
- Merge, decay, sort, band-shuffle, and author-cap all run in Go on at most `forYouCandidateCap = 800` candidates.
- The new-count endpoint runs a single indexed `COUNT(*)` capped at `LIMIT 1001`, cached 30 s. Even at 30 s polling per user the DB load is bounded.
- Pagination remains offset-based for now. If active feed sessions routinely cross deeper than ~page 50, switch to cursor pagination encoding `(score, thread_id)` to avoid offset cost.
