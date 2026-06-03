---
summary: The author-chosen control over who can reply to a thread — EVERYONE (default), FOLLOWING, or MUTUAL.
last_synced: 2026-06-03
source:
  - frontend/app/about/threads/page.tsx
  - backend/domain/model/thread/thread.go
tags: [concept, threads, permissions]
---

# Reply permissions

When composing a [[threads|thread]], the author chooses who can reply:

- **`EVERYONE`** — the default; anyone can reply.
- **`FOLLOWING`** — only your followers can reply.
- **`MUTUAL`** — only people you mutually follow can reply.

This is set per-thread at compose time. See [[threads]] for the rest of the posting flow.
