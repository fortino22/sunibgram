# For You Feed Algorithm

`GET /threads/` returns a personalized For You feed for authenticated users. Unauthenticated requests keep using the existing chronological public feed.

## Feed Score

Threads have a precomputed `feed_score` column updated every 5 minutes for content from the last 7 days:

```text
(likes + 3 * comments + 5 * reposts + 4 * bookmarks)
/ (1 + (age_hours ^ 1.5))
```

Threads older than 7 days are reset to `feed_score = 0`.

## Candidate Sources

The repository fetches up to 100 recent post/repost candidates from each source, ordered by `feed_score DESC`:

- Following posts: users followed by the viewer, weight `1.0`.
- Community posts: communities where the viewer is a member, weight `0.9`.
- Faculty posts: authors in the viewer's faculty, excluding the viewer, weight `0.7`.
- Global trending: top public posts from Redis key `feed:trending:global`, weight `0.5`.

Final score is `feed_score * source_weight`. If the same thread appears in multiple sources, the highest final score is kept. Reposts are only allowed from the following source.

## Diversity

After sorting by final score, the feed applies an author cap of 2 posts per author per page result. This prevents one author from dominating a page while preserving high-scoring content order.

## Caching

- Per-user For You page cache: `feed:foryou:{userID}:{page}:{limit}`, TTL 120 seconds.
- Global trending cache: `feed:trending:global`, TTL 5 minutes.

The background scheduler refreshes `feed_score` and the global trending cache every 5 minutes.

## Scaling Notes

The feed uses a stored `feed_score` column and a partial index on recent rows to avoid recomputing ranking during requests. Redis absorbs repeated reads for both global trending and per-user pages. Candidate generation stays bounded by limiting each source to 100 rows before merge, dedupe, diversity capping, and pagination in Go.
