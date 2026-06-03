---
summary: Sunibgram's on-platform AI analyst — user-facing thread summaries plus platform-side event review and daily/weekly reports.
last_synced: 2026-06-03
source:
  - frontend/app/about/fox/page.tsx
tags: [entity, ai, fox]
---

# Fox

Fox is Sunibgram's on-platform AI analyst. It has two faces: a **user-facing** summary
tool and a **platform-side** analyst role. It runs on the [[hermes]] toolset.

## Thread summaries (user-facing)
Open a [[threads|thread]] → generate a **Fox** summary of the post plus its replies. Your
summary history lives at `/summaries` (the Fox-agent tab) — see [[user-routes]].

## Analyst role (platform-side)
- **Event review:** Fox reviews every **official** [[events|event]] and returns
  approve / decline / ask-for-changes with a confidence score and a written reason.
  Clear-cut low-risk events can auto-approve; anything risky is flagged for a human.
  Low-confidence or borderline calls **never** auto-approve — they route to an admin
  with Fox's notes.
- **Daily morning brief:** events decided, bugs, top feature requests.
- **Weekly story:** trending hashtags, top engaging users, safety notes — built from live
  activity.

Fox also powers the **For You** feed transparency page (`/about/algorithm`) — see
[[feed-ranking]].

See [[hermes]] for the toolset Fox runs on, and [[about]] for the AI/automation overview.
