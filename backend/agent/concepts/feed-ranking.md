---
summary: How the For You feed is ordered — a quiet interest score from reaction weights, time decay, fresh slots, a fairness cap, and per-visit reshuffle.
last_synced: 2026-06-03
source:
  - frontend/app/about/algorithm/page.tsx
tags: [concept, feed, ranking, algorithm]
---

# Feed / For You ranking

How the **For You** tab on the home feed (`/`, see [[user-routes]]) is ordered. The
transparency page is `/about/algorithm`; [[fox]] powers it.

- **Interest score:** posts get a quiet score from reactions, freshness, and source (your
  posts, people you follow, your [[communities]], campus-wide trending).
- **Reaction weights:** reposts count most, then bookmarks/comments, then shares, then a
  like (baseline); replies are nudged down vs original posts.
- **Time decay:** strongest right after posting, fading over hours; posts older than
  ~a week leave For You.
- **Fresh slots:** a couple of just-posted threads (~last 15 min) get a reserved spot up
  top, newest first.
- **Fairness cap:** no single account can dominate — limited posts near the top and overall.
- **Reshuffle:** similar-interest posts are gently shuffled so the feed varies per visit.
- **"New posts" banner:** appears while scrolling when fresh threads arrive; tap to slide
  them in without losing your place.

The other feed tabs (Following, Faculty, Community) are described in [[threads]].
