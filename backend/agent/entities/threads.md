---
summary: Posts, replies, reposts, quotes, and polls — the core unit of Sunibgram, with author-controlled reply permissions and three feed views.
last_synced: 2026-06-03
source:
  - frontend/app/about/threads/page.tsx
  - backend/domain/model/thread/thread.go
tags: [entity, threads, posting]
---

# Threads & posts

The core unit of Sunibgram. To create a post or reply, use the composer on the home feed
(`/`) or the reply box inside a thread — there is no separate "new post" page. Open a
single thread at `/thread/:id` (see [[user-routes]]).

## Thread types
`POST`, `REPLY`, `REPOST`.

## Reply permissions
The author chooses who can reply — `EVERYONE` (default), `FOLLOWING`, or `MUTUAL`. Full
detail in [[reply-permissions]].

## Reposts & quotes
One-tap repost, or quote with your own comment (the quoted thread embeds inline). Undo
with a tap; original credit stays with the author.

## Polls
Add a poll to a **post-type** thread (not replies/reposts); live results. Caps —
options, label length, question length, duration, option-image size — are in [[limits]].

## More
- `@mention` autocomplete while composing.
- Edit / delete / pin your own threads any time.
- **Feeds:** switch between **For You**, **Following**, and **Faculty** in one tap. The
  faculty feed shows posts only from verified students in your faculty. Ranking for For
  You is in [[feed-ranking]].

Text and media caps (500 chars, max 4 media, allowed types) live in [[limits]].
Visibility of public vs community threads is in [[privacy-model]].
