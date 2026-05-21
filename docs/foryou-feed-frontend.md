# For You Feed — Frontend Integration Guide

This document describes how the Next.js frontend should integrate with the new For You feed backend (Phase 1 algorithm + Phase 2 endpoints). The backend has shipped these changes; the frontend currently works against them without modification, but the UX features below require small additions.

---

## TL;DR

Three endpoints power the feed UX:

| Endpoint | Method | Purpose |
|---|---|---|
| `/threads/` | `GET` | Paginated feed (existing, unchanged contract) |
| `/threads/feed/refresh` | `POST` | Reset the user's feed ordering seed |
| `/threads/feed/new-count` | `GET` | Count of new threads since `since_id` |

Without any frontend changes, the feed already feels dynamic: each new session sees a reshuffled ranking. The two new endpoints are required only if the frontend wants:

1. A "Show N new posts" banner (X / Threads-style).
2. A user-controlled pull-to-refresh gesture that produces a different order.

---

## How the Backend Behaves Now

The For You feed (`GET /threads/`) returns a deterministic ordering for a given session "seed". The seed:

- Is minted server-side, per user, and stored in Redis (10 minute sliding TTL).
- Stays the same across paginated requests, so `page=1`, `page=2`, ... return a stable, contiguous list with no duplicates.
- Expires automatically after 10 minutes of feed inactivity → next request mints a new one.
- Can be explicitly dropped via `POST /threads/feed/refresh`.

Inside a single seed window, the algorithm:

1. Merges 4 candidate sources (own posts, global trending, following, communities).
2. Applies log-based time decay so older posts drift down naturally.
3. Sorts by score and shuffles **within score bands** so trending posts stay near the top but exact order varies between sessions.
4. Caps to 20 posts per author.
5. Slices the page.

You do not need to think about any of this on the frontend — just call the endpoint, get a page, render it. The "feels fresh on refresh" behavior is automatic.

---

## Endpoint 1 — `GET /threads/`

**Unchanged.** Same request shape, same response DTO (`PaginatedFeedResponse`).

```ts
// existing TanStack Query hook, no changes needed
const { data, fetchNextPage } = useInfiniteQuery({
  queryKey: ['threads', 'feed'],
  queryFn: ({ pageParam = 1 }) =>
    api.get<PaginatedFeedResponse>(`/threads/?page=${pageParam}&limit=20`),
  getNextPageParam: (lastPage, all) =>
    all.length * 20 < lastPage.meta.total ? all.length + 1 : undefined,
})
```

Pages are stable within a 10-minute window of activity. Each subsequent request refreshes the TTL.

---

## Endpoint 2 — `POST /threads/feed/refresh`

Invalidates the caller's feed seed so the next `GET /threads/` returns a freshly shuffled order. No request body, no response body.

```http
POST /threads/feed/refresh
Authorization: Bearer <jwt>
→ 204 No Content
```

### When to call it

- On a **pull-to-refresh gesture** at the top of the feed.
- When the user taps the "Show N new posts" banner (see Endpoint 3).
- Optionally: when the user re-enters the feed screen after extended idle time (e.g. >5 min in another tab/screen).

Do **not** call it on every focus event or periodic timer — the heuristic effect of repeatedly invalidating the seed will undo pagination stability.

### Recommended wrapper

```ts
async function refreshForYouFeed() {
  await api.post('/threads/feed/refresh')
  await queryClient.invalidateQueries({ queryKey: ['threads', 'feed'] })
}
```

After `POST /refresh` the frontend **must** also refetch the feed (`invalidateQueries`) — the POST alone only resets the seed in Redis; the rendered list will not change until a new GET is issued.

---

## Endpoint 3 — `GET /threads/feed/new-count`

Returns the number of feed-eligible threads created after a given thread id. Used to drive the "Show N new posts" banner.

```http
GET /threads/feed/new-count?since_id=12345
Authorization: Bearer <jwt>
→ 200 OK
{
  "count": 47
}
```

- `since_id` — the largest `thread_id` currently rendered at the top of the user's feed.
- `count` — number of new threads, capped at `1000` (display as `"999+"` in UI).
- Result is cached server-side for 30 seconds per `(userID, since_id)` pair, so polling is cheap.

### Polling strategy

Use TanStack Query with `refetchInterval`:

```ts
const topThread = feedPages?.[0]?.data?.[0]
const sinceID = topThread?.thread_id

const { data: newCountData } = useQuery({
  queryKey: ['threads', 'feed', 'new-count', sinceID],
  queryFn: () => api.get<{ count: number }>(`/threads/feed/new-count?since_id=${sinceID}`),
  enabled: !!sinceID,
  refetchInterval: 30_000,           // poll every 30s
  refetchIntervalInBackground: false, // pause when tab not focused
})

const newCount = newCountData?.count ?? 0
```

### Banner component

```tsx
{newCount > 0 && (
  <button
    type="button"
    onClick={handleShowNew}
    className="sticky top-0 z-10 w-full bg-primary py-2 text-center text-sm font-medium text-primary-foreground"
  >
    Show {newCount >= 1000 ? '999+' : newCount} new posts
  </button>
)}
```

### Handler

```ts
async function handleShowNew() {
  await api.post('/threads/feed/refresh')                      // mint new seed
  await queryClient.invalidateQueries(['threads', 'feed'])     // refetch page 1+
  await queryClient.invalidateQueries(['threads', 'feed', 'new-count']) // reset banner
  feedListRef.current?.scrollTo({ top: 0, behavior: 'smooth' })
}
```

---

## Pull-to-refresh

If the feed view is in a scroll container and you support pull-to-refresh (mobile gesture or library like `react-pull-to-refresh`):

```ts
async function onPullToRefresh() {
  await refreshForYouFeed()
}
```

This is functionally identical to tapping the new-posts banner, just triggered by gesture.

---

## State the frontend should track

| State | Source | Purpose |
|---|---|---|
| `feedPages` | `useInfiniteQuery(['threads', 'feed'])` | Rendered list |
| `sinceID` | `feedPages[0].data[0].thread_id` | Newest visible thread id |
| `newCount` | `useQuery(['threads', 'feed', 'new-count', sinceID])` | Banner number |

There is **no client-side seed to manage** — the seed lives entirely in Redis. The frontend's only job is to call `/refresh` at the right moments and refetch.

---

## Lifecycle Walkthrough

```
T = 0s     App opens → GET /threads/?page=1
           - Backend mints seed S1 (Redis TTL 10min)
           - Returns 20 threads ordered by S1
           - Frontend stores sinceID = data[0].thread_id (= 1000)

T = 5s     User scrolls → GET /threads/?page=2
           - Backend reuses S1 (TTL refreshed)
           - Returns next 20 threads (no duplicates)

T = 30s    Background poll → GET /threads/feed/new-count?since_id=1000
           - Backend returns { count: 12 }
           - Frontend renders banner "Show 12 new posts"

T = 60s    Poll fires again → { count: 47 }
           - Banner updates to "Show 47 new posts"

T = 2min   User taps banner OR pulls to refresh
           - POST /threads/feed/refresh           (seed S1 deleted)
           - queryClient.invalidateQueries(['threads', 'feed'])
           - GET /threads/?page=1
             - Backend mints seed S2 → reshuffled order returned
             - data[0].thread_id = 1047
           - Frontend sets sinceID = 1047, scrolls to top
           - Banner clears (new-count query refetches → 0)

T = 12min  User idle, no feed requests
           - Seed S2 TTL expires server-side
           - Next GET will mint S3 → automatically fresh ordering

T = ...    User presses browser refresh / hard reload
           - All frontend state lost
           - GET /threads/?page=1 again
           - Backend seed (if still within TTL) → SAME order user saw before reload
           - This is expected: F5 should not surprise users with new content
```

---

## What NOT to do

- ❌ Do not call `POST /threads/feed/refresh` on every focus/visibility event.
- ❌ Do not poll `/feed/new-count` faster than every 15s.
- ❌ Do not store the seed in localStorage/sessionStorage — it is server-side only.
- ❌ Do not assume a refresh endpoint call alone updates the UI — you must refetch the feed query afterward.
- ❌ Do not pass `since_id` from an older page (e.g. middle of the list) — always use the **top** thread id.

---

## Edge Cases

| Case | Expected behavior |
|---|---|
| User logs out / switches account | Seed is per-`userID`; new user gets their own seed automatically. |
| User opens app on two tabs | Same seed → identical order across tabs. ✓ |
| User scrolls past minute boundary | Seed TTL is sliding (refreshed on every read), so as long as they keep scrolling within 10 min, order stays stable. |
| User pulls to refresh while at page 5 | `invalidateQueries` resets pagination cursor; user is returned to page 1 of the fresh order. |
| New-count query returns 0 | Banner hidden — `newCount > 0` guard. |
| `since_id` is 0 (no threads loaded yet) | Skip the query (`enabled: !!sinceID`). |
| Backend returns 5xx on `/refresh` | Treat as soft failure — still invalidate local feed query so user sees a refetch attempt. |

---

## Quick Component Sketch

```tsx
'use client'

import { useInfiniteQuery, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRef } from 'react'

export function ForYouFeed() {
  const queryClient = useQueryClient()
  const listRef = useRef<HTMLDivElement>(null)

  const feedQuery = useInfiniteQuery({
    queryKey: ['threads', 'feed'],
    queryFn: ({ pageParam = 1 }) =>
      api.get<PaginatedFeedResponse>(`/threads/?page=${pageParam}&limit=20`),
    getNextPageParam: (last, all) =>
      all.flatMap(p => p.data).length < last.meta.total ? all.length + 1 : undefined,
    initialPageParam: 1,
  })

  const sinceID = feedQuery.data?.pages[0]?.data[0]?.thread_id

  const newCountQuery = useQuery({
    queryKey: ['threads', 'feed', 'new-count', sinceID],
    queryFn: () =>
      api.get<{ count: number }>(`/threads/feed/new-count?since_id=${sinceID}`),
    enabled: !!sinceID,
    refetchInterval: 30_000,
    refetchIntervalInBackground: false,
  })

  const newCount = newCountQuery.data?.count ?? 0

  async function handleShowNew() {
    await api.post('/threads/feed/refresh')
    await queryClient.invalidateQueries({ queryKey: ['threads', 'feed'] })
    await queryClient.invalidateQueries({ queryKey: ['threads', 'feed', 'new-count'] })
    listRef.current?.scrollTo({ top: 0, behavior: 'smooth' })
  }

  return (
    <div ref={listRef} className="relative overflow-y-auto">
      {newCount > 0 && (
        <button
          onClick={handleShowNew}
          className="sticky top-0 z-10 w-full bg-primary py-2 text-center text-sm font-medium text-primary-foreground"
        >
          Show {newCount >= 1000 ? '999+' : newCount} new posts
        </button>
      )}
      {feedQuery.data?.pages.flatMap(p => p.data).map(thread => (
        <ThreadCard key={thread.thread_id} thread={thread} />
      ))}
    </div>
  )
}
```

---

## Backend Reference (FYI only)

For debugging, the new Redis key shape:

```
feed:foryou:user:{userID}:seed              — current feed seed, sliding 10 min TTL
feed:foryou:user:{userID}:src:own           — 30s TTL
feed:foryou:user:{userID}:src:following     — 60s TTL
feed:foryou:user:{userID}:src:community     — 60s TTL
feed:foryou:user:{userID}:newcount:{since}  — 30s TTL
feed:trending:global                        — 2 min TTL
```

You should never read or write these from the frontend — they are listed only so backend engineers can verify behavior in Redis during integration testing.
